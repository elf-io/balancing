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

// LocalRedirectPoliciesGetter has a method to return a LocalRedirectPolicyInterface.
// A group's client should implement this interface.
type LocalRedirectPoliciesGetter interface {
	LocalRedirectPolicies() LocalRedirectPolicyInterface
}

// LocalRedirectPolicyInterface has methods to work with LocalRedirectPolicy resources.
type LocalRedirectPolicyInterface interface {
	Create(ctx context.Context, localRedirectPolicy *v1beta1.LocalRedirectPolicy, opts v1.CreateOptions) (*v1beta1.LocalRedirectPolicy, error)
	Update(ctx context.Context, localRedirectPolicy *v1beta1.LocalRedirectPolicy, opts v1.UpdateOptions) (*v1beta1.LocalRedirectPolicy, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, localRedirectPolicy *v1beta1.LocalRedirectPolicy, opts v1.UpdateOptions) (*v1beta1.LocalRedirectPolicy, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.LocalRedirectPolicy, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.LocalRedirectPolicyList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.LocalRedirectPolicy, err error)
	LocalRedirectPolicyExpansion
}

// localRedirectPolicies implements LocalRedirectPolicyInterface
type localRedirectPolicies struct {
	*gentype.ClientWithList[*v1beta1.LocalRedirectPolicy, *v1beta1.LocalRedirectPolicyList]
}

// newLocalRedirectPolicies returns a LocalRedirectPolicies
func newLocalRedirectPolicies(c *BalancingV1beta1Client) *localRedirectPolicies {
	return &localRedirectPolicies{
		gentype.NewClientWithList[*v1beta1.LocalRedirectPolicy, *v1beta1.LocalRedirectPolicyList](
			"localredirectpolicies",
			c.RESTClient(),
			scheme.ParameterCodec,
			"",
			func() *v1beta1.LocalRedirectPolicy { return &v1beta1.LocalRedirectPolicy{} },
			func() *v1beta1.LocalRedirectPolicyList { return &v1beta1.LocalRedirectPolicyList{} }),
	}
}
