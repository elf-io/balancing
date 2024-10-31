// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package policyId

import (
	"errors"
	balancing "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/types"
	"github.com/elf-io/balancing/pkg/utils"
)

var PolicyErrorMissId = errors.New("policy miss id")
var PolicyErrorInvalidId = errors.New("policy has invalid id")

func GetBalancingPolicyValidity(policy *balancing.BalancingPolicy) (uint32, error) {
	if idStr, ok := policy.Annotations[types.AnnotationServiceID]; ok {
		if id, e := utils.StringToUint32(idStr); e != nil {
			return 0, PolicyErrorInvalidId
		} else {
			return id, nil
		}
	}
	return 0, PolicyErrorMissId
}

func GetLocalRedirectPolicyValidity(policy *balancing.LocalRedirectPolicy) (uint32, error) {
	if idStr, ok := policy.Annotations[types.AnnotationServiceID]; ok {
		if id, e := utils.StringToUint32(idStr); e != nil {
			return 0, PolicyErrorInvalidId
		} else {
			return id, nil
		}
	}
	return 0, PolicyErrorMissId

}
