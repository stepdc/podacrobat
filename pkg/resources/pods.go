package resources

import (
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/apis/core/v1/helper/qos"
	policyvb1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func Evictable(pod *v1.Pod) bool {
	// check local mount & DaemonSet only
	if pod == nil {
		return false
	}

	for _, vol := range pod.Spec.Volumes {
		if vol.EmptyDir != nil || vol.HostPath != nil {
			return false
		}
	}

	for _, ref := range pod.ObjectMeta.GetOwnerReferences() {
		if ref.Kind == "DaemonSet" {
			return false
		}
	}

	return true
}

func FilterEvictablePods(pods []*v1.Pod) []*v1.Pod {
	var ret []*v1.Pod
	for _, pod := range pods {
		if !Evictable(pod) {
			continue
		}
		ret = append(ret, pod)
	}
	return ret
}

func IsBestEffortPod(pod *v1.Pod) bool {
	return qos.GetPodQOS(pod) == v1.PodQOSBestEffort
}

func IsBurstablePod(pod *v1.Pod) bool {
	return qos.GetPodQOS(pod) == v1.PodQOSBurstable
}

func IsGuaranteedPod(pod *v1.Pod) bool {
	return qos.GetPodQOS(pod) == v1.PodQOSGuaranteed
}

func Evict(cli clientset.Interface, pod *v1.Pod) (error) {
	ev := policyvb1.Eviction{
		TypeMeta: metav1.TypeMeta{
			Kind: "Eviction",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: pod.Name,
			Namespace: pod.Namespace,
		},
		DeleteOptions: &metav1.DeleteOptions{},
	}
	err := cli.PolicyV1beta1().Evictions(ev.Namespace).Evict(&ev)
	if err != nil {
		return fmt.Errorf("evict %q failed: %v",  err)
	}
	return nil
}