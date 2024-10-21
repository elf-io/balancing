// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

// +kubebuilder:object:generate=true
// +groupName=balancing.elf.io

package v1beta1

import (
	"github.com/elf-io/balancing/pkg/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: types.ApiGroup, Version: types.ApiVersion}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
