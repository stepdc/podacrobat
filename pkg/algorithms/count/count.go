package count

import (
	"fmt"
	"strconv"

	"github.com/stepdc/podacrobat/cmd/app/config"
	v1 "k8s.io/api/core/v1"
)

// sample algo for test
type PodCountAlgo struct{}

func (pac *PodCountAlgo) NeedReschedule(groupedPods map[string][]*v1.Pod, cfg config.Config) (bool, error) {
	l, err := strconv.Atoi(cfg.LowerThreshold)
	if err != nil {
		return false, err
	}
	u, err := strconv.Atoi(cfg.UpperThreshold)
	if err != nil {
		return false, err
	}

	var lmatched, umatched bool
	for node, pods := range groupedPods {
		if err != nil {
			return false, fmt.Errorf("check %q if need reschedule failed :%v", node, err)
		}

		if len(pods) <= l {
			lmatched = true
		}
		if len(pods) >= u {
			umatched = true
		}

		if lmatched && umatched {
			return true, nil
		}
	}

	return false, nil
}
