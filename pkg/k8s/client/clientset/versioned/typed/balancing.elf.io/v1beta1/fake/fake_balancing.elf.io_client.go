// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1beta1 "github.com/elf-io/balancing/pkg/k8s/client/clientset/versioned/typed/balancing.elf.io/v1beta1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeBalancingV1beta1 struct {
	*testing.Fake
}

func (c *FakeBalancingV1beta1) BalancingPolicies() v1beta1.BalancingPolicyInterface {
	return &FakeBalancingPolicies{c}
}

func (c *FakeBalancingV1beta1) LocalRedirectPolicies() v1beta1.LocalRedirectPolicyInterface {
	return &FakeLocalRedirectPolicies{c}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeBalancingV1beta1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
