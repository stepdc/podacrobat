package count

import (
	"fmt"
	"log"

	v1 "k8s.io/api/core/v1"

	"github.com/stepdc/podacrobat/cmd/app/config"
	"github.com/stepdc/podacrobat/pkg/resources"

	clientset "k8s.io/client-go/kubernetes"
)

type countOptions struct {
	lower, upper int
}

// simple algo for test
type PodCountAlgo struct {
	option countOptions
}

func NewPodCountAlgo(cfg config.Config) *PodCountAlgo {
	return &PodCountAlgo{
		option: countOptions{
			lower: cfg.IdleCountThreshold,
			upper: cfg.EvictCountThreshold,
		},
	}
}

func (pac *PodCountAlgo) Run(cli clientset.Interface, nodePods map[string]resources.NodeInfoWithPods) error {
	needRun := pac.NeedReschedule(nodePods)
	if !needRun {
		log.Println("cluster is balanced")
		return nil
	}
	lower, load := pac.ClassifyNodes(nodePods)
	return pac.Evict(cli, lower, load)
}

func (pac *PodCountAlgo) NeedReschedule(nodePods map[string]resources.NodeInfoWithPods) bool {
	if len(nodePods) <= 1 {
		return false
	}
	var lmatched, umatched bool
	for _, node := range nodePods {
		if len(node.Pods) <= pac.option.lower {
			lmatched = true
		}
		if len(node.Pods) >= pac.option.upper {
			umatched = true
		}

		if lmatched && umatched {
			return true
		}
	}

	return false
}

func (pac *PodCountAlgo) ClassifyNodes(nodePods map[string]resources.NodeInfoWithPods) (map[string]resources.NodeInfoWithPods, map[string]resources.NodeInfoWithPods) {
	idleNodes := make(map[string]resources.NodeInfoWithPods)
	evictNodes := make(map[string]resources.NodeInfoWithPods)

	for nodeName, info := range nodePods {
		if info.Node.Spec.Unschedulable {
			continue
		}
		if len(info.Pods) <= pac.option.lower {
			idleNodes[nodeName] = info
		}
		if len(info.Pods) >= pac.option.upper {
			evictNodes[nodeName] = info
		}
	}

	return idleNodes, evictNodes
}

func (pac *PodCountAlgo) Evict(cli clientset.Interface, idleNodes, evictNodes map[string]resources.NodeInfoWithPods) error {
	total := totalPodCapacity(idleNodes, pac.option.lower)
	shouldEvictTotal := mostEvictCount(evictNodes, pac.option.upper)

	var refsSet map[string]struct{}
	// evict BestEffort & Burstable pods only
	for _, info := range evictNodes {
		if total <= 0 || shouldEvictTotal <= 0 {
			return nil
		}
		bePods := info.BestEffortPods()
		var evictedBePods []*v1.Pod
		var err error
		evictedBePods, refsSet, err = resources.EvictPods(cli, bePods, refsSet)
		if err != nil {
			err = fmt.Errorf("evict pods failed: %v", err)
			log.Print(err)
			return err
		}
		log.Printf("evict %v BestEffort level for node %v", len(evictedBePods), info.Node.Name)
		total -= len(evictedBePods)
		shouldEvictTotal -= len(evictedBePods)

		if total <= 0 || shouldEvictTotal <= 0 {
			return nil
		}
		buPods := info.BurstablePods()
		var evictedBuPods []*v1.Pod
		evictedBuPods, refsSet, err = resources.EvictPods(cli, buPods, nil)
		if err != nil {
			err = fmt.Errorf("evict pods failed: %v", err)
			log.Print(err)
			return err
		}
		log.Printf("evict %v Burstable level for node %v", len(evictedBuPods), info.Node.Name)
		total -= len(evictedBuPods)
		shouldEvictTotal -= len(evictedBuPods)
	}

	return nil
}

func totalPodCapacity(nodePods map[string]resources.NodeInfoWithPods, threshold int) int {
	var ret int
	for _, node := range nodePods {
		c := threshold - len(node.Pods)
		if c > 0 {
			ret += c
		}
	}

	return ret
}

func mostEvictCount(nodePods map[string]resources.NodeInfoWithPods, threshold int) int {
	var ret int
	for _, node := range nodePods {
		pods := resources.FilterEvictablePods(node.Pods)
		c := len(pods) - threshold
		if c > 0 {
			ret += c
		}
	}
	return ret
}
