// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

// rbac marker:
// https://github.com/kubernetes-sigs/controller-tools/blob/master/pkg/rbac/parser.go
// https://book.kubebuilder.io/reference/markers/rbac.html

// for crd
// +kubebuilder:rbac:groups=balancing.elf.io,resources=localredirectpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=balancing.elf.io,resources=localredirectpolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=balancing.elf.io,resources=balancingpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=balancing.elf.io,resources=balancingpolicies/status,verbs=get;update;patch

// for k8s object, check 'kubectl api-resources -o wide'
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=nodes;services;pods;namespaces,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="discovery.k8s.io",resources=endpointslices,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=create;get;update
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete

package v1beta1
