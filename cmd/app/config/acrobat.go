package config

import (
	"github.com/spf13/pflag"

	clientset "k8s.io/client-go/kubernetes"
)

type PodAcrobat struct {
	Config
	Client clientset.Interface
}

func (pa *PodAcrobat) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&pa.Policy, "policy", string(PodsCount), "nodes filter policy(use \"podscount\" for test)")
	fs.StringVar(&pa.LowerThreshold, "lowerthreshold", "30", "lower threshold")
	fs.StringVar(&pa.UpperThreshold, "upperthreshold", "50", "upper threshold")
}
