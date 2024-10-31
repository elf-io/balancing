// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"context"
	"fmt"
	balancingv1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/policyId"
	"github.com/elf-io/balancing/pkg/types"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type webhookRedirect struct {
	l *zap.Logger
}

var _ webhook.CustomDefaulter = (*webhookRedirect)(nil)

// mutating webhook
func (s *webhookRedirect) Default(ctx context.Context, obj runtime.Object) error {
	bp, ok := obj.(*balancingv1beta1.LocalRedirectPolicy)
	if !ok {
		s.l.Sugar().Errorf("expected a LocalRedirectPolicy but got a %T", obj)
		return apierrors.NewBadRequest("failed to get LocalRedirectPolicy")
	}

	logger := s.l.Named("LocalRedirectMutating").With(
		zap.String("crdName", bp.Name),
	)

	if _, e := policyId.GetLocalRedirectPolicyValidity(bp); e == policyId.PolicyErrorMissId {
		if idStr, e := policyId.PolicyIdManagerHandler.GeneratePolicyId(bp.Name); e != nil {
			msg := fmt.Sprintf("failed to generate id: %v ", e)
			logger.Sugar().Errorf(msg)
			return fmt.Errorf(msg)
		} else {
			bp.Annotations[types.AnnotationServiceID] = idStr
			logger.Sugar().Infof("add service Id=%s to policy", idStr)
		}
	} else if e == policyId.PolicyErrorInvalidId {
		idStr := bp.Annotations[types.AnnotationServiceID]
		logger.Sugar().Errorf("policy has an invalid Id: %s=%s", types.AnnotationServiceID, idStr)
	} else {
		idStr := bp.Annotations[types.AnnotationServiceID]
		logger.Sugar().Debugf("valid Id: %s=%s", types.AnnotationServiceID, idStr)
	}

	return nil
}

func (s *webhookRedirect) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	bp, ok := obj.(*balancingv1beta1.LocalRedirectPolicy)
	if !ok {
		s.l.Sugar().Errorf("expected a LocalRedirectPolicy but got a %T", obj)
		return nil, apierrors.NewBadRequest("failed to get LocalRedirectPolicy")
	}

	logger := s.l.Named("LocalRedirectMutating").With(
		zap.String("crdName", bp.Name),
	)

	id, e := policyId.GetLocalRedirectPolicyValidity(bp)
	if e == policyId.PolicyErrorMissId {
		msg := fmt.Sprintf("policy miss service Id, the mutating webhook went wrong ")
		logger.Sugar().Errorf(msg)
		return nil, fmt.Errorf(msg)
	} else if e == policyId.PolicyErrorInvalidId {
		msg := fmt.Sprintf("policy has an invalid Id in annotation ")
		logger.Sugar().Errorf(msg)
		return nil, fmt.Errorf(msg)
	} else {
		if e := policyId.PolicyIdManagerHandler.SavePolicyId(bp.Name, fmt.Sprintf("%d", id)); e != nil {
			msg := fmt.Sprintf("failed to save id: %v ", e)
			logger.Sugar().Errorf(msg)
		} else {
			logger.Sugar().Infof("save service id: %d", id)
		}
	}

	return nil, nil
}

func (s *webhookRedirect) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newBp, ok := newObj.(*balancingv1beta1.LocalRedirectPolicy)
	if !ok {
		s.l.Sugar().Errorf("expected a LocalRedirectPolicy but got a %T", newObj)
		return nil, apierrors.NewBadRequest("failed to get LocalRedirectPolicy")
	}
	oldBp := oldObj.(*balancingv1beta1.LocalRedirectPolicy)

	logger := s.l.Named("LocalRedirectMutating").With(
		zap.String("crdName", oldBp.Name),
	)

	if !reflect.DeepEqual(newBp.Spec, oldBp.Spec) {
		msg := fmt.Sprintf("policy is not allowed to update the spec")
		logger.Sugar().Errorf(msg)
		return nil, fmt.Errorf(msg)
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (s *webhookRedirect) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	bp, ok := obj.(*balancingv1beta1.LocalRedirectPolicy)
	if !ok {
		s.l.Sugar().Errorf("expected a LocalRedirectPolicy but got a %T", obj)
		return nil, apierrors.NewBadRequest("failed to get LocalRedirectPolicy")
	}

	logger := s.l.Named("LocalRedirectMutating").With(
		zap.String("crdName", bp.Name),
	)
	logger.Sugar().Debugf("policy deleted")

	return nil, nil
}
