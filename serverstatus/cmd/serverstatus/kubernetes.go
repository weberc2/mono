package main

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func newClientset() (*kubernetes.Clientset, error) {
	// 1. Try in-cluster config
	config, err := rest.InClusterConfig()
	if err == nil {
		return kubernetes.NewForConfig(config)
	}

	// 2. Fallback to kubeconfig (local dev)
	home, err2 := os.UserHomeDir()
	if err2 != nil {
		return nil, fmt.Errorf("cannot find home directory for kubeconfig: %w, and also not in cluster: %v", err2, err)
	}

	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err3 := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err3 != nil {
		return nil, fmt.Errorf("cannot build kubeconfig: %v (in-cluster error: %v)", err3, err)
	}

	return kubernetes.NewForConfig(config)
}
