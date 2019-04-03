package util

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stepdc/podacrobat/pkg/resources"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/stepdc/podacrobat/cmd/app/config"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clienttesting "k8s.io/client-go/testing"
)

func TestUtil(t *testing.T) {
	cfg := config.Config{
		Policy:                config.NodesLoad,
		CpuUtilEvictThreshold: 50,
		CpuUtilIdleThreshold:  20,
		MemUtilEvictThreshold: 50,
		MemUtilIdleThreshold:  20,
	}
	algo := NewCpuMemUtilAlgo(cfg)

	pod1 := genTestPod("test-pod-1", "test-node-1", "ref1", 100, 100)
	pod2 := genTestPod("test-pod-1", "test-node-1", "ref2", 100, 100)
	pod3 := genTestPod("test-pod-1", "test-node-1", "ref3", 100, 100)
	pod4 := genTestPod("test-pod-1", "test-node-1", "ref4", 100, 100)
	pod5 := genTestPod("test-pod-1", "test-node-1", "ref5", 100, 100)
	pod6 := genTestPod("test-pod-1", "test-node-1", "ref6", 100, 100)
	node1Pods := append([]*v1.Pod(nil), pod1, pod2, pod3, pod4, pod5, pod6)

	pod7 := genTestPod("test-pod-1", "test-node-1", "ref7", 100, 100)
	node2Pods := []*v1.Pod{pod7}
	// pod8 := genTestPod("test-pod-1", "test-node-1", "ref8", 100, 100)

	node1 := genTestNode("test-node-1", 1000, 1000)
	node2 := genTestNode("test-node-2", 1000, 1000)

	fakeCli := &fake.Clientset{}
	fakeCli.Fake.AddReactor("list", "pods", func(action clienttesting.Action) (bool, runtime.Object, error) {
		list := action.(clienttesting.ListAction)
		fieldString := list.GetListRestrictions().Fields.String()
		if strings.Contains(fieldString, "node-1") {
			return true, &v1.PodList{Items: []v1.Pod{*pod1, *pod2, *pod3, *pod4, *pod5, *pod6}}, nil
		} else if strings.Contains(fieldString, "node-2") {
			return true, &v1.PodList{Items: []v1.Pod{*pod7}}, nil
		}
		return true, nil, fmt.Errorf("list failed: %v", list)
	})
	fakeCli.Fake.AddReactor("get", "nodes", func(action clienttesting.Action) (bool, runtime.Object, error) {
		get := action.(clienttesting.GetAction)
		getName := get.GetName()
		if strings.Contains(getName, "test-node-1") {
			return true, node1, nil
		} else if strings.Contains(getName, "test-node-2") {
			return true, node2, nil
		}
		return true, nil, fmt.Errorf("get node %q failed", getName)
	})

	nodePods := map[string]resources.NodeInfoWithPods{
		node1.Name: resources.NodeInfoWithPods{Node: node1, Pods: node1Pods},
		node2.Name: resources.NodeInfoWithPods{Node: node2, Pods: node2Pods},
	}
	err := algo.Run(fakeCli, nodePods)
	if err != nil {
		t.Error(err)
	}
}

func genTestPod(name, nodeName, refName string, cpu, mem int) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
			SelfLink:  fmt.Sprintf("/api/v1/namespaces/default/pods/%s", name),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{},
						Limits:   v1.ResourceList{},
					},
				},
			},
			NodeName: nodeName,
		},
	}
	if cpu >= 0 {
		pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = *resource.NewMilliQuantity(int64(cpu), resource.DecimalSI)
	}
	if mem >= 0 {
		pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = *resource.NewQuantity(int64(mem), resource.DecimalSI)
	}

	pod.OwnerReferences = []metav1.OwnerReference{metav1.OwnerReference{APIVersion: "v1", Kind: "ReplicaSet", Name: refName}}

	return pod
}

func genTestNode(name string, cpu, mem int) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     name,
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", name),
			Labels:   map[string]string{},
		},
		Status: v1.NodeStatus{
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(int64(cpu), resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(int64(mem), resource.DecimalSI),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(int64(cpu), resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(int64(mem), resource.DecimalSI),
			},
			Phase: v1.NodeRunning,
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue},
			},
		},
	}

}
