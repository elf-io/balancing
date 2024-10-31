// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"context"
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpfWriter"
	balancing "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/policyId"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcilerBalancing struct {
	client client.Client
	l      *zap.Logger
	writer ebpfWriter.EbpfWriter
}

func (s *ReconcilerBalancing) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := s.l.With(
		zap.String("policy", req.NamespacedName.Name),
	)
	res := reconcile.Result{}

	// Fetch the ReplicaSet from the cache
	rs := &balancing.BalancingPolicy{}
	err := s.client.Get(ctx, req.NamespacedName, rs)
	if errors.IsNotFound(err) {
		logger.Sugar().Debugf("policy %v has been deleted", req.NamespacedName.Name)
		if e := s.writer.DeleteBalancingByPolicy(logger, req.NamespacedName.Name); e != nil {
			logger.Sugar().Errorf("%v", e)
		}
		return res, nil
	} else if err != nil {
		logger.Sugar().Debugf("could not fetch %v: %+v", req.NamespacedName.Name, err)
		return res, fmt.Errorf("could not fetch: %+v", err)
	}

	if _, e := policyId.GetBalancingPolicyValidity(rs); e != nil {
		logger.Sugar().Errorf("localRedirect policy %v is invalid: %v", req.NamespacedName.Name, e)
		return res, nil
	}
	logger.Sugar().Debugf("reconcile: balancing policy %v", req.NamespacedName.Name)

	if e := s.writer.UpdateBalancingByPolicy(logger, rs); e != nil {
		logger.Sugar().Errorf("%v", e)
	}

	return res, nil

}
