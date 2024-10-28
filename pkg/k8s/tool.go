// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package k8s

import discoveryv1 "k8s.io/api/discovery/v1"

func GetEndpointSliceOwnerName(t *discoveryv1.EndpointSlice) string {
	// for default/kubernetes ，there is no owner
	edsName := t.Namespace + "/" + t.Name
	if len(t.OwnerReferences) > 0 {
		edsName = t.Namespace + "/" + t.OwnerReferences[0].Name
	}
	return edsName
}
