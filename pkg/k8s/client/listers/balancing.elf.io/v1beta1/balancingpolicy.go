// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
)

// BalancingPolicyLister helps list BalancingPolicies.
// All objects returned here must be treated as read-only.
type BalancingPolicyLister interface {
	// List lists all BalancingPolicies in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.BalancingPolicy, err error)
	// Get retrieves the BalancingPolicy from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1beta1.BalancingPolicy, error)
	BalancingPolicyListerExpansion
}

// balancingPolicyLister implements the BalancingPolicyLister interface.
type balancingPolicyLister struct {
	listers.ResourceIndexer[*v1beta1.BalancingPolicy]
}

// NewBalancingPolicyLister returns a new BalancingPolicyLister.
func NewBalancingPolicyLister(indexer cache.Indexer) BalancingPolicyLister {
	return &balancingPolicyLister{listers.New[*v1beta1.BalancingPolicy](indexer, v1beta1.Resource("balancingpolicy"))}
}
