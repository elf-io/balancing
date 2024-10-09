// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

// !!!!!! crd marker:
// kubectl get  如何打印
// https://github.com/kubernetes-sigs/controller-tools/blob/master/pkg/crd/markers/crd.go
// https://book.kubebuilder.io/reference/markers/crd.html
// 字段验证
// https://github.com/kubernetes-sigs/controller-tools/blob/master/pkg/crd/markers/validation.go
// https://book.kubebuilder.io/reference/markers/crd-validation.html

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RedirectMode string

const (
	RedirectModePodEndpoint = RedirectMode("podEndpoint")
	RedirectModeNodePort    = RedirectMode("nodePort")
	RedirectModeNodeProxy   = RedirectMode("nodeProxy")
)

type ServiceEndpoint struct {
	// Namespace is the Kubernetes service namespace.
	// The service namespace must match the namespace of the parent Local
	// Redirect Policy.  For Cluster-wide Local Redirect Policy, this
	// can be any namespace.
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// Name is the name of a destination Kubernetes service that identifies traffic
	// to be redirected.
	// The service type needs to be ClusterIP.
	//
	// +kubebuilder:validation:Required
	ServiceName string `json:"serviceName"`

	// RedirectMode defines the destination IP
	//
	// +kubebuilder:validation:Enum=podEndpoint;nodeProxy;nodePort
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=podEndpoint
	RedirectMode RedirectMode `json:"redirectMode"`

	// ToPorts is a list of destination service L4 ports with protocol for
	// traffic to be redirected. If not specified, traffic for all the service
	// ports will be redirected.
	// When multiple ports are specified, the ports must be named.
	//
	// +kubebuilder:validation:Optional
	ToPorts []PortInfo `json:"toPorts,omitempty"`
}

type BalancingBackend struct {
	// AddressEndpoint is a tuple {IP, port, protocol} where the traffic will be redirected.
	//
	// +kubebuilder:validation:OneOf
	AddressEndpoint []*AddressEndpoint `json:"addressEndpoint,omitempty"`

	// serviceEndpoint are pods where the traffic will be redirected.
	//
	// +kubebuilder:validation:OneOf
	ServiceEndpoint *ServiceEndpoint `json:"serviceEndpoint,omitempty"`
}

// ----------------------------

type BalancingSpec struct {
	// enable this policy
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Enabled *bool `json:"enabled,omitempty"`

	// BalancingFrontend specifies frontend configuration to redirect traffic from.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf", message="frontend is immutable"
	BalancingFrontend RedirectFrontend `json:"frontend"`

	// BalancingBackend specifies backend configuration to redirect traffic to.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf", message="backend is immutable"
	BalancingBackend BalancingBackend `json:"backend"`
}

type BalancingStatus struct {
	Enabled bool `json:"enabled,omitempty"`
}

// adds a column to "kubectl get" output for this CRD
// https://github.com/kubernetes-sigs/controller-tools/blob/main/pkg/crd/markers/crd.go#L195
//
// +kubebuilder:resource:categories={elf},path="balancingpolicys",singular="balancingpolicy",scope="Cluster",shortName={bl}
// +kubebuilder:printcolumn:JSONPath=".spec.frontend.serviceMatcher.serviceName",description="serviceName",name="serviceName",type=string
// +kubebuilder:printcolumn:JSONPath=".spec.frontend.serviceMatcher.namespace",description="namespace",name="namespace",type=string
// +kubebuilder:printcolumn:JSONPath=".spec.enabled",description="enabled",name="enabled",type=boolean
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +genclient
// +genclient:nonNamespaced
type BalancingPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   BalancingSpec   `json:"spec,omitempty"`
	Status BalancingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type BalancingPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []BalancingPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BalancingPolicy{}, &BalancingPolicyList{})
}
