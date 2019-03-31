package acrobat

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

var ErrListNodesTimeout = errors.New("list nodes timeout")

func ListNodes(ctx context.Context, cli clientset.Interface) ([]*v1.Node, error) {
	select {
	case <-ctx.Done():
		return nil, ErrListNodesTimeout
	default:
		return listNodes(cli)
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
