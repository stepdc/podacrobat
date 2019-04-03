package resources

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

var ErrListNodesTimeout = errors.New("list nodes timeout")
var ErrListNodePodsTimeout = errors.New("list node pods timeout")

func ListNodes(ctx context.Context, cli clientset.Interface) ([]*v1.Node, error) {
	type result struct {
		nodes []*v1.Node
		err   error
	}
	c := make(chan result, 1)

	go func(ctx context.Context) {
		nodes, err := listNodes(cli)
		result := result{nodes, err}
		select {
		case c <- result:
		case <-ctx.Done():
		}
	}(ctx)

	select {
	case <-ctx.Done():
		return nil, ErrListNodesTimeout
	case ret := <-c:
		return ret.nodes, ret.err
	}
}

func listNodes(cli clientset.Interface) ([]*v1.Node, error) {
	items, err := cli.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes failed: %v", err)
	}
	if items == nil || len(items.Items) == 0 {
		return nil, nil
	}

	var ret []*v1.Node
	for _, node := range items.Items {
		node := node
		ret = append(ret, &node)
	}

	return filterReady(ret), nil
}

func filterReady(nodes []*v1.Node) []*v1.Node {
	var ret []*v1.Node

	for _, node := range nodes {
		var notReady bool
		for _, status := range node.Status.Conditions {
			if status.Type != v1.NodeReady {
				notReady = true
				break
			}
		}
		if notReady {
			continue
		}
		ret = append(ret, node)
	}

	return ret
}

func GroupPodsByNode(cli clientset.Interface, nodes []*v1.Node) (map[string]NodeInfoWithPods, error) {
	ret := make(map[string]NodeInfoWithPods)

	for _, node := range nodes {
		pods, err := listNodePods(cli, node)
		if err != nil {
			return nil, fmt.Errorf("list pods on node %q faild: %v", node.Name, err)
		}
		ret[node.Name] = NodeInfoWithPods{Node: node, Pods: pods}
	}

	return ret, nil
}

func listNodePods(cli clientset.Interface, node *v1.Node) ([]*v1.Pod, error) {
	sstr := fmt.Sprintf(`spec.nodeName=%s,status.phase=%s||status.phase\=%s`, "node1", "Pending", "Running")
	selector, err := fields.ParseSelector(sstr)
	if err != nil {
		return nil, fmt.Errorf("parse pod selector failed: %v", err)
	}

	pods, err := cli.CoreV1().Pods("").List(metav1.ListOptions{FieldSelector: selector.String()})

	var ret []*v1.Pod
	for _, pod := range pods.Items {
		pod := pod
		ret = append(ret, &pod)
	}

	return ret, nil
}

type NodeInfoWithPods struct {
	Node *v1.Node
	Pods []*v1.Pod
}

func (n *NodeInfoWithPods) BestEffortPods() []*v1.Pod {
	var ret []*v1.Pod
	for _, pod := range n.Pods {
		if !IsBestEffortPod(pod) {
			continue
		}
		ret = append(ret, pod)
	}
	return ret
}

func (n *NodeInfoWithPods) BurstablePods() []*v1.Pod {
	var ret []*v1.Pod
	for _, pod := range n.Pods {
		if !IsBurstablePod(pod) {
			continue
		}
		ret = append(ret, pod)
	}
	return ret
}

func IsIdleNode(usage, capacity v1.ResourceList, cpuThreshold, memThreshold float64) bool {
	cpu, mem := UsagePercentage(usage, capacity)

	if cpu > cpuThreshold || mem > memThreshold {
		return false
	}

	return true
}

func IsEvictNode(usage, capacity v1.ResourceList, cpuThreshold, memThreshold float64) bool {
	cpu, mem := UsagePercentage(usage, capacity)

	if cpu < cpuThreshold || mem < memThreshold {
		return false
	}

	return true
}

func UsagePercentage(usage, capacity v1.ResourceList) (float64, float64) {
	cpuUsag := usage[v1.ResourceCPU]
	memUsage := usage[v1.ResourceMemory]
	cpuCapacity := capacity[v1.ResourceCPU]
	memCapacity := capacity[v1.ResourceMemory]
	cpu := float64(cpuUsag.MilliValue()) * 100 / float64(cpuCapacity.MilliValue())
	mem := float64(memUsage.Value()) * 100 / float64(memCapacity.Value())
	return cpu, mem
}
