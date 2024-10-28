// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"context"
	balancingv1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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

	logger.Sugar().Infof("mutating")

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

	logger.Sugar().Infof("ValidateCreate")

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

	logger.Sugar().Infof("ValidateUpdate %+v", newBp)

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (s *webhookRedirect) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}
