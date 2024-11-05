// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeBalancingPolicies implements BalancingPolicyInterface
type FakeBalancingPolicies struct {
	Fake *FakeBalancingV1beta1
}

var balancingpoliciesResource = v1beta1.SchemeGroupVersion.WithResource("balancingpolicies")

var balancingpoliciesKind = v1beta1.SchemeGroupVersion.WithKind("BalancingPolicy")

// Get takes name of the balancingPolicy, and returns the corresponding balancingPolicy object, and an error if there is any.
func (c *FakeBalancingPolicies) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.BalancingPolicy, err error) {
	emptyResult := &v1beta1.BalancingPolicy{}
	obj, err := c.Fake.
		Invokes(testing.NewRootGetActionWithOptions(balancingpoliciesResource, name, options), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.BalancingPolicy), err
}

// List takes label and field selectors, and returns the list of BalancingPolicies that match those selectors.
func (c *FakeBalancingPolicies) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.BalancingPolicyList, err error) {
	emptyResult := &v1beta1.BalancingPolicyList{}
	obj, err := c.Fake.
		Invokes(testing.NewRootListActionWithOptions(balancingpoliciesResource, balancingpoliciesKind, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.BalancingPolicyList{ListMeta: obj.(*v1beta1.BalancingPolicyList).ListMeta}
	for _, item := range obj.(*v1beta1.BalancingPolicyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested balancingPolicies.
func (c *FakeBalancingPolicies) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchActionWithOptions(balancingpoliciesResource, opts))
}

// Create takes the representation of a balancingPolicy and creates it.  Returns the server's representation of the balancingPolicy, and an error, if there is any.
func (c *FakeBalancingPolicies) Create(ctx context.Context, balancingPolicy *v1beta1.BalancingPolicy, opts v1.CreateOptions) (result *v1beta1.BalancingPolicy, err error) {
	emptyResult := &v1beta1.BalancingPolicy{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(balancingpoliciesResource, balancingPolicy, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.BalancingPolicy), err
}

// Update takes the representation of a balancingPolicy and updates it. Returns the server's representation of the balancingPolicy, and an error, if there is any.
func (c *FakeBalancingPolicies) Update(ctx context.Context, balancingPolicy *v1beta1.BalancingPolicy, opts v1.UpdateOptions) (result *v1beta1.BalancingPolicy, err error) {
	emptyResult := &v1beta1.BalancingPolicy{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateActionWithOptions(balancingpoliciesResource, balancingPolicy, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.BalancingPolicy), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeBalancingPolicies) UpdateStatus(ctx context.Context, balancingPolicy *v1beta1.BalancingPolicy, opts v1.UpdateOptions) (result *v1beta1.BalancingPolicy, err error) {
	emptyResult := &v1beta1.BalancingPolicy{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceActionWithOptions(balancingpoliciesResource, "status", balancingPolicy, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.BalancingPolicy), err
}

// Delete takes name of the balancingPolicy and deletes it. Returns an error if one occurs.
func (c *FakeBalancingPolicies) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(balancingpoliciesResource, name, opts), &v1beta1.BalancingPolicy{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeBalancingPolicies) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionActionWithOptions(balancingpoliciesResource, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.BalancingPolicyList{})
	return err
}

// Patch applies the patch and returns the patched balancingPolicy.
func (c *FakeBalancingPolicies) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.BalancingPolicy, err error) {
	emptyResult := &v1beta1.BalancingPolicy{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(balancingpoliciesResource, name, pt, data, opts, subresources...), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.BalancingPolicy), err
}