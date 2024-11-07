// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/elf-io/balancing/pkg/ebpfWriter"
	balancing "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/policyId"
	"github.com/elf-io/balancing/pkg/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
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

func CheckLocalNodeSelected(ctx context.Context, c client.Client, config *balancing.PolicyConfig, nodeName string) (bool, error) {
	if config == nil {
		return true, nil
	}
	if config != nil && config.NodeLabelSelector == nil {
		return true, nil
	}
	if config.NodeLabelSelector.MatchLabels == nil && len(config.NodeLabelSelector.MatchExpressions) == 0 {
		return true, nil
	}

	// get local node
	localNodeName := types.AgentConfig.LocalNodeName
	node := &corev1.Node{}
	err := c.Get(ctx, client.ObjectKey{Name: localNodeName}, node)
	if err != nil {
		if errors.IsNotFound(err) {
			if config.EnableOutCluster {
				// for hosts out of the cluster
				return true, nil
			} else {
				// for hosts out of the cluster, take effect by default
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to get local node %v: %+v", localNodeName, err)
	}

	// match
	labelSelector, err := metav1.LabelSelectorAsSelector(config.NodeLabelSelector)
	if err != nil {
		return false, fmt.Errorf("failed to get LabelSelectorAsSelector: %v", err)
	}
	if !labelSelector.Matches(labels.Set(node.Labels)) {
		return false, nil
	}

	return true, nil
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
	logger.Sugar().Debugf("reconcile: balancing policy %v", req.NamespacedName.Name)

	if ok, err := CheckLocalNodeSelected(ctx, s.client, rs.Spec.Config, types.AgentConfig.LocalNodeName); err != nil {
		logger.Sugar().Errorf("%v", err)
		return res, err
	} else {
		if !ok {
			logger.Sugar().Infof("policy does not select local node, ignore it")
			return res, nil
		}
		logger.Sugar().Debugf("policy selects local node")
	}

	if _, e := policyId.GetBalancingPolicyValidity(rs); e != nil {
		logger.Sugar().Errorf("localRedirect policy %v is invalid: %v", req.NamespacedName.Name, e)
		return res, nil
	}

	if e := s.writer.UpdateBalancingByPolicy(logger, rs); e != nil {
		logger.Sugar().Errorf("%v", e)
	}

	return res, nil

}
