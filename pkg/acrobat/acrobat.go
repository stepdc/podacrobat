package acrobat

import (
	"context"
	"fmt"
	"time"

	"github.com/stepdc/podacrobat/cmd/app/config"
	"github.com/stepdc/podacrobat/pkg/algorithms/count"
	"github.com/stepdc/podacrobat/pkg/resources"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const defaultTimeout = 30 * time.Second

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

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	avaliableNodes, err := resources.ListNodes(ctx, pa.Client)
	if err != nil {
		return fmt.Errorf("filter nodes failed: %v", err)
	}
	if len(avaliableNodes) == 0 {
		// noready nodes
		return nil
	}

	groupedPods, err := resources.GroupPodsByNode(cli, avaliableNodes)
	if err != nil {
		return err
	}

	// TODO: use algo interface here
	algo := count.PodCountAlgo{}
	if resched, err := algo.NeedReschedule(groupedPods, pa.Config); err != nil || !resched {
		return err
	}

	return nil
}
