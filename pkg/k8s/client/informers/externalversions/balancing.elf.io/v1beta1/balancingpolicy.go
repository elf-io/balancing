// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

// Code generated by informer-gen. DO NOT EDIT.

package v1beta1

import (
	"context"
	time "time"

	balancingelfiov1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	versioned "github.com/elf-io/balancing/pkg/k8s/client/clientset/versioned"
	internalinterfaces "github.com/elf-io/balancing/pkg/k8s/client/informers/externalversions/internalinterfaces"
	v1beta1 "github.com/elf-io/balancing/pkg/k8s/client/listers/balancing.elf.io/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// BalancingPolicyInformer provides access to a shared informer and lister for
// BalancingPolicies.
type BalancingPolicyInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1beta1.BalancingPolicyLister
}

type balancingPolicyInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewBalancingPolicyInformer constructs a new informer for BalancingPolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewBalancingPolicyInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredBalancingPolicyInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredBalancingPolicyInformer constructs a new informer for BalancingPolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredBalancingPolicyInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.BalancingV1beta1().BalancingPolicies().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.BalancingV1beta1().BalancingPolicies().Watch(context.TODO(), options)
			},
		},
		&balancingelfiov1beta1.BalancingPolicy{},
		resyncPeriod,
		indexers,
	)
}

func (f *balancingPolicyInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredBalancingPolicyInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *balancingPolicyInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&balancingelfiov1beta1.BalancingPolicy{}, f.defaultInformer)
}

func (f *balancingPolicyInformer) Lister() v1beta1.BalancingPolicyLister {
	return v1beta1.NewBalancingPolicyLister(f.Informer().GetIndexer())
}