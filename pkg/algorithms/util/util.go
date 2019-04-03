package util

import (
	"fmt"
	"log"
	"strconv"

	v1 "k8s.io/api/core/v1"

	"github.com/stepdc/podacrobat/cmd/app/config"
	"github.com/stepdc/podacrobat/pkg/resources"

	clientset "k8s.io/client-go/kubernetes"
)

type cmuOption struct {
	cpuEvictThreshold, cpuIdleThreshold float64
	memEvictThreshold, memIdleThreshold float64
}

type CpuMemUtilAlgo struct {
	cmuOption
}

func NewCpuMemUtilAlgo(cfg config.Config) *CpuMemUtilAlgo {
	ce, _ := strconv.ParseFloat(cfg.CpuUtilEvictThreshold, 64)
	ci, _ := strconv.ParseFloat(cfg.CpuUtilIdleThreshold, 64)
	me, _ := strconv.ParseFloat(cfg.MemUtilEvictThreshold, 64)
	mi, _ := strconv.ParseFloat(cfg.MemUtilIdleThreshold, 64)

	algo := CpuMemUtilAlgo{
		cmuOption{
			cpuEvictThreshold: ce,
			cpuIdleThreshold:  ci,
			memEvictThreshold: me,
			memIdleThreshold:  mi,
		},
	}

	return &algo
}

func (cmu *CpuMemUtilAlgo) Run(cli clientset.Interface, nodePods map[string]resources.NodeInfoWithPods) error {
	idles, evicts := cmu.ClassifyNodes(nodePods)
	if len(idles) == 0 || len(evicts) == 0 {
		log.Printf("cluster is balanced")
		return nil
	}

	return cmu.Evict(cli, idles, evicts)
}

func (cmu *CpuMemUtilAlgo) ClassifyNodes(nodePods map[string]resources.NodeInfoWithPods) (map[string]resources.NodeInfoWithPods, map[string]resources.NodeInfoWithPods) {
	idle := make(map[string]resources.NodeInfoWithPods)
	evict := make(map[string]resources.NodeInfoWithPods)

	for nname, info := range nodePods {
		if info.Node.Spec.Unschedulable {
			continue
		}
		podsUsage := resources.PodsCpuMemRequest(info.Pods)
		nodeCapacity := info.Node.Status.Capacity
		if resources.IsIdleNode(podsUsage, nodeCapacity, cmu.cpuIdleThreshold, cmu.memIdleThreshold) {
			idle[nname] = info
			continue
		}
		if resources.IsEvictNode(podsUsage, nodeCapacity, cmu.cpuEvictThreshold, cmu.memEvictThreshold) {
			evict[nname] = info
		}
	}
	return idle, evict
}

func (cmu *CpuMemUtilAlgo) Evict(cli clientset.Interface, idles map[string]resources.NodeInfoWithPods, evicts map[string]resources.NodeInfoWithPods) error {
	cpuTargetThreshold := targetThreshold(cmu.cpuIdleThreshold, cmu.cpuEvictThreshold)
	memTargetThreshold := targetThreshold(cmu.memIdleThreshold, cmu.memEvictThreshold)
	totalCpu, totalMem := totalIdleCapacity(idles, cpuTargetThreshold, memTargetThreshold)
	if totalCpu <= 0 || totalMem <= 0 {
		log.Printf("no room for pods to reschedule, quit now")
		return nil
	}

	refs := make(map[string]struct{})
	var err error
	for nodeName, info := range evicts {
		targetEvictCpu, targetEvictMem := evictCapacity(info, cpuTargetThreshold, memTargetThreshold)
		if targetEvictCpu > totalCpu {
			targetEvictCpu = totalCpu
		}
		if targetEvictMem > totalMem {
			targetEvictMem = totalMem
		}

		bePods := info.BestEffortPods()
		buPods := info.BurstablePods()
		var evicted []*v1.Pod
		evicted, refs, err = resources.EvictTargetQuantityPods(cli, append(bePods, buPods...), targetEvictCpu, targetEvictMem, refs)
		if err != nil {
			return fmt.Errorf("evict pods for node %q failed: %v", nodeName, err)
		}
		evictedResource := resources.PodsCpuMemRequest(evicted)
		evictedCpu := evictedResource[v1.ResourceCPU]
		evictedMem := evictedResource[v1.ResourceMemory]
		totalCpu -= float64(evictedCpu.MilliValue())
		totalMem -= float64(evictedMem.Value())
		if totalCpu < 0 || totalMem < 0 {
			break
		}
	}

	return nil
}

func targetThreshold(idle, evict float64) float64 {
	return idle + (evict-idle)/2
}

func totalIdleCapacity(idles map[string]resources.NodeInfoWithPods, cpuThreshold, memThreshold float64) (float64, float64) {
	var totalCpu, totalMem float64
	for _, info := range idles {
		usage := resources.PodsCpuMemRequest(info.Pods)
		capacity := info.Node.Status.Capacity

		cpuUsedPer, memUsedPer := resources.UsagePercentage(usage, capacity)
		availableCpuPer := cpuThreshold - cpuUsedPer
		if availableCpuPer > 0 {
			totalCpu += availableCpuPer * float64(capacity.Cpu().MilliValue()) / 100
		}
		availableMemPer := memThreshold - memUsedPer
		if availableMemPer > 0 {
			totalMem += availableMemPer * float64(capacity.Memory().Value()) / 100
		}
	}
	return totalCpu, totalMem
}

func evictCapacity(nodeInfo resources.NodeInfoWithPods, cpuThreshold, memThreshold float64) (float64, float64) {
	var cpu, mem float64
	usage := resources.PodsCpuMemRequest(nodeInfo.Pods)
	capacity := nodeInfo.Node.Status.Capacity
	cpuUsedPer, memUsedPer := resources.UsagePercentage(usage, capacity)
	evictCpuPer := cpuUsedPer - cpuThreshold
	if evictCpuPer > 0 {
		cpu = evictCpuPer * float64(capacity.Cpu().MilliValue()) / 100
	}
	evictMemPer := memUsedPer - memThreshold
	if evictMemPer > 0 {
		mem = evictMemPer * float64(capacity.Memory().Value()) / 100
	}
	return cpu, mem
}
