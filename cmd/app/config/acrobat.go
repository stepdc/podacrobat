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
	fs.StringVar(&pa.Policy, "policy", PodsCount, "nodes filter policy(use \"podscount\" for test)")
	fs.IntVar(&pa.IdleCountThreshold, "lowerthreshold", 30, "lower threshold")
	fs.IntVar(&pa.EvictCountThreshold, "upperthreshold", 50, "upper threshold")
	fs.Float64Var(&pa.CpuUtilIdleThreshold, "util-cpu-idle-threshold", 20, "util cpu idle threshold")
	fs.Float64Var(&pa.CpuUtilEvictThreshold, "util-cpu-evict-threshold", 60, "util cpu evict threshold")
	fs.Float64Var(&pa.MemUtilIdleThreshold, "util-memory-idle-threshold", 20, "util memory idle threshold")
	fs.Float64Var(&pa.MemUtilEvictThreshold, "util-memory-evict-threshold", 60, "util memory evict threshold")
}
