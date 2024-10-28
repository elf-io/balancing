// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"context"
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpfWriter"
	balancing "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/types"
	"github.com/elf-io/balancing/pkg/utils"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcilerRedirect struct {
	client client.Client
	l      *zap.Logger
	writer ebpfWriter.EbpfWriter
}

func (s *ReconcilerRedirect) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	CheckPolicyValidity := func(policy *balancing.LocalRedirectPolicy) error {
		if idStr, ok := policy.Annotations[types.AnnotationServiceID]; ok {
			if _, e := utils.StringToUint32(idStr); e != nil {
				return fmt.Errorf("policy %s has an invalid serviceID annotation %s, skip", idStr)
			}
		} else {
			return fmt.Errorf("policy %s miss serviceID annotation, skip", policy.Name)
		}
		return nil
	}

	logger := s.l.With(
		zap.String("policy", req.NamespacedName.Name),
	)
	res := reconcile.Result{}

	// Fetch the ReplicaSet from the cache
	rs := &balancing.LocalRedirectPolicy{}
	err := s.client.Get(ctx, req.NamespacedName, rs)
	if errors.IsNotFound(err) {
		logger.Sugar().Debugf("policy %v has been deleted", req.NamespacedName.Name)
		if e := s.writer.DeleteRedirectByPolicy(logger, req.NamespacedName.Name); e != nil {
			logger.Sugar().Errorf("%v", e)
		}
		return res, nil
	} else if err != nil {
		logger.Sugar().Debugf("could not fetch %v: %+v", req.NamespacedName.Name, err)
		return res, fmt.Errorf("could not fetch: %+v", err)
	}
	logger.Sugar().Debugf("reconcile: LocalRedirectPolicy policy %s", req.NamespacedName.Name)

	if e := CheckPolicyValidity(rs); e != nil {
		logger.Sugar().Errorf("localRedirect policy %v is invalid: %v", req.NamespacedName.Name, e)
		return res, nil
	}

	if e := s.writer.UpdateRedirectByPolicy(logger, rs); e != nil {
		logger.Sugar().Errorf("%v", e)
	}

	return res, nil
}
