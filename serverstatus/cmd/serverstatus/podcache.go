package main

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listerscorev1 "k8s.io/client-go/listers/core/v1"
)

// Wrapper type that exposes a list function backed by an informer/lister.
type podCache struct {
	factory   informers.SharedInformerFactory
	podLister listerscorev1.PodLister
}

func newPodCache(clientset *kubernetes.Clientset) (p podCache) {
	p.factory = informers.NewSharedInformerFactory(clientset, 30*time.Second)
	p.podLister = p.factory.Core().V1().Pods().Lister()
	return
}

// Start begins watching pods and populating the cache.
func (p *podCache) Start(ctx context.Context) {
	p.factory.Start(ctx.Done())
	// Wait for informer sync
	p.factory.WaitForCacheSync(ctx.Done())
}

// List all pods across all namespaces (cached).
func (p *podCache) ListPods() ([]*v1.Pod, error) {
	return p.podLister.List(labels.Everything())
}
