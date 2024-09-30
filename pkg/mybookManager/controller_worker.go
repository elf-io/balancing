// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package mybookManager

import (
	"context"
	crd "github.com/elf-io/elf/pkg/k8s/apis/elf.elf.io/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

func (s *informerHandler) syncHandler(ctx context.Context, obj *crd.Mybook) error {
	if obj == nil {
		return nil
	}
	logger := s.logger.Named("worker")

	// 通过 clientset 向 api server 实时获取最新数据
	// old, err := s.k8sclient.ElfV1().Mybooks().Get(ctx, obj.Name, metav1.GetOptions{})
	// 获取最新cache中的数据（cache中的数据有延时风险）
	old, err := s.crdlister.Get(obj.Name)
	if err != nil {
		logger.Warn("failed to get " + obj.Name)
		if apierrors.IsNotFound(err) {
			// not found ,no retry
			return nil
		}
		// retry later
		return err
	}
	logger.Info("handle " + obj.Name)

	newone := old.DeepCopy()
	newone.Status.TotalIPCount = 100

	if !reflect.DeepEqual(old, newone) {
		if _, err := s.k8sclient.ElfV1().Mybooks().UpdateStatus(ctx, newone, metav1.UpdateOptions{}); err != nil {
			// if conflicted, queue will retry it later
			return err
		}
		logger.Info("succeed to update " + obj.Name)

		// generate crd event
		s.eventRecord.Eventf(newone, corev1.EventTypeNormal, "modified Mybook", "crd event, new mybook %v", newone.Name)

	}

	return nil
}
