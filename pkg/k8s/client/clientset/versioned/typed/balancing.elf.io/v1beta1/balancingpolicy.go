// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	"context"

	v1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	scheme "github.com/elf-io/balancing/pkg/k8s/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// BalancingPoliciesGetter has a method to return a BalancingPolicyInterface.
// A group's client should implement this interface.
type BalancingPoliciesGetter interface {
	BalancingPolicies() BalancingPolicyInterface
}

// BalancingPolicyInterface has methods to work with BalancingPolicy resources.
type BalancingPolicyInterface interface {
	Create(ctx context.Context, balancingPolicy *v1beta1.BalancingPolicy, opts v1.CreateOptions) (*v1beta1.BalancingPolicy, error)
	Update(ctx context.Context, balancingPolicy *v1beta1.BalancingPolicy, opts v1.UpdateOptions) (*v1beta1.BalancingPolicy, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.BalancingPolicy, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.BalancingPolicyList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.BalancingPolicy, err error)
	BalancingPolicyExpansion
}

// balancingPolicies implements BalancingPolicyInterface
type balancingPolicies struct {
	*gentype.ClientWithList[*v1beta1.BalancingPolicy, *v1beta1.BalancingPolicyList]
}

// newBalancingPolicies returns a BalancingPolicies
func newBalancingPolicies(c *BalancingV1beta1Client) *balancingPolicies {
	return &balancingPolicies{
		gentype.NewClientWithList[*v1beta1.BalancingPolicy, *v1beta1.BalancingPolicyList](
			"balancingpolicies",
			c.RESTClient(),
			scheme.ParameterCodec,
			"",
			func() *v1beta1.BalancingPolicy { return &v1beta1.BalancingPolicy{} },
			func() *v1beta1.BalancingPolicyList { return &v1beta1.BalancingPolicyList{} }),
	}
}
