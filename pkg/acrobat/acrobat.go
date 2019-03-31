package acrobat

import (
	"fmt"

	"github.com/stepdc/podacrobat/cmd/app/config"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Run(pa *config.PodAcrobat) error {
	// incluster supported only
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("could not generated incluster configuration for kubernetes: %v", err)
	}
	cli, err := clientset.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("build client failed: %v", err)
	}
	pa.Client = cli

	return nil
}
