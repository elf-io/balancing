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

// FakeLocalRedirectPolicies implements LocalRedirectPolicyInterface
type FakeLocalRedirectPolicies struct {
	Fake *FakeBalancingV1beta1
}

var localredirectpoliciesResource = v1beta1.SchemeGroupVersion.WithResource("localredirectpolicies")

var localredirectpoliciesKind = v1beta1.SchemeGroupVersion.WithKind("LocalRedirectPolicy")

// Get takes name of the localRedirectPolicy, and returns the corresponding localRedirectPolicy object, and an error if there is any.
func (c *FakeLocalRedirectPolicies) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.LocalRedirectPolicy, err error) {
	emptyResult := &v1beta1.LocalRedirectPolicy{}
	obj, err := c.Fake.
		Invokes(testing.NewRootGetActionWithOptions(localredirectpoliciesResource, name, options), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.LocalRedirectPolicy), err
}

// List takes label and field selectors, and returns the list of LocalRedirectPolicies that match those selectors.
func (c *FakeLocalRedirectPolicies) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.LocalRedirectPolicyList, err error) {
	emptyResult := &v1beta1.LocalRedirectPolicyList{}
	obj, err := c.Fake.
		Invokes(testing.NewRootListActionWithOptions(localredirectpoliciesResource, localredirectpoliciesKind, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.LocalRedirectPolicyList{ListMeta: obj.(*v1beta1.LocalRedirectPolicyList).ListMeta}
	for _, item := range obj.(*v1beta1.LocalRedirectPolicyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested localRedirectPolicies.
func (c *FakeLocalRedirectPolicies) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchActionWithOptions(localredirectpoliciesResource, opts))
}

// Create takes the representation of a localRedirectPolicy and creates it.  Returns the server's representation of the localRedirectPolicy, and an error, if there is any.
func (c *FakeLocalRedirectPolicies) Create(ctx context.Context, localRedirectPolicy *v1beta1.LocalRedirectPolicy, opts v1.CreateOptions) (result *v1beta1.LocalRedirectPolicy, err error) {
	emptyResult := &v1beta1.LocalRedirectPolicy{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(localredirectpoliciesResource, localRedirectPolicy, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.LocalRedirectPolicy), err
}

// Update takes the representation of a localRedirectPolicy and updates it. Returns the server's representation of the localRedirectPolicy, and an error, if there is any.
func (c *FakeLocalRedirectPolicies) Update(ctx context.Context, localRedirectPolicy *v1beta1.LocalRedirectPolicy, opts v1.UpdateOptions) (result *v1beta1.LocalRedirectPolicy, err error) {
	emptyResult := &v1beta1.LocalRedirectPolicy{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateActionWithOptions(localredirectpoliciesResource, localRedirectPolicy, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.LocalRedirectPolicy), err
}

// Delete takes name of the localRedirectPolicy and deletes it. Returns an error if one occurs.
func (c *FakeLocalRedirectPolicies) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(localredirectpoliciesResource, name, opts), &v1beta1.LocalRedirectPolicy{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeLocalRedirectPolicies) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionActionWithOptions(localredirectpoliciesResource, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.LocalRedirectPolicyList{})
	return err
}

// Patch applies the patch and returns the patched localRedirectPolicy.
func (c *FakeLocalRedirectPolicies) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.LocalRedirectPolicy, err error) {
	emptyResult := &v1beta1.LocalRedirectPolicy{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(localredirectpoliciesResource, name, pt, data, opts, subresources...), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.LocalRedirectPolicy), err
}
