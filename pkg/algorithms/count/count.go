package count

import (
	"github.com/golang/glog"
	"github.com/stepdc/podacrobat/pkg/resources"
	"strconv"

	"github.com/stepdc/podacrobat/cmd/app/config"
	clientset "k8s.io/client-go/kubernetes"
)

type countOptions struct {
	lower, upper int
}

// simple algo for test
type PodCountAlgo struct{
	option countOptions
}

func NewPodCountAlgo(cfg config.Config) *PodCountAlgo {
	l, _ := strconv.Atoi(cfg.LowerThreshold)
	u, _:= strconv.Atoi(cfg.UpperThreshold)
	return &PodCountAlgo{
	option: countOptions{
		lower: l,
		upper: u,
	},
}
}

func (pac *PodCountAlgo) NeedReschedule(nodePods map[string]resources.NodeInfoWithPods, cfg config.Config) (bool, error) {
	var lmatched, umatched bool
	for _, node := range nodePods {

		if len(node.Pods) <= pac.option.lower {
			lmatched = true
		}
		if len(node.Pods) >= pac.option.upper {
			umatched = true
		}

		if lmatched && umatched {
			return true, nil
		}
	}

	return false, nil
}

func (pac *PodCountAlgo) ClassifyNodes(nodePods map[string]resources.NodeInfoWithPods,
	cfg config.Config) (map[string]resources.NodeInfoWithPods, map[string]resources.NodeInfoWithPods) {

	lowerNodes := make(map[string]resources.NodeInfoWithPods)
	loadNodes := make(map[string]resources.NodeInfoWithPods)

	for nodeName, info := range nodePods {
		if len(info.Pods) <= pac.option.lower {
			lowerNodes[nodeName] = info
		}
		if len(info.Pods) >= pac.option.upper {
			loadNodes[nodeName] = info
		}
	}

	return lowerNodes, loadNodes
}

func (pac *PodCountAlgo) Evict(cli clientset.Interface,lowerNodes, loadNodes map[string]resources.NodeInfoWithPods) error {
	total := totalPodCapacity(lowerNodes, pac.option.lower)

	// evict BestEffort & Burstable pods only
	for _, info := range loadNodes {
		bePods := info.BestEffortPods()
		for _, pod := range bePods {
			err := resources.Evict(cli, pod)
			if err != nil {
				glog.Error(err)
				continue
			}
			glog.Infof("%q evicted", pod.Name)
			total--
			if total == 0 {
				return nil
			}
		}

		buPods := info.BurstablePods()
		for _, pod := range buPods {
			err := resources.Evict(cli, pod)
			if err != nil {
				glog.Error(err)
				continue
			}
			glog.Infof("%q evicted", pod.Name)
			total--
			if total == 0 {
				return nil
			}
		}
	}

	return nil
}

func totalPodCapacity(nodePods map[string]resources.NodeInfoWithPods, threshold int) int {
	var ret int
	for _, node := range nodePods {
		v := threshold - len(node.Pods)
		if v > 0 {
			ret += v
		}
	}

	return ret
}
