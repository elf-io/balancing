package cmd

import (
	"context"
	"fmt"
	balancing "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reconcilerRedirect struct {
	client client.Client
	l      *zap.Logger
}

func (s *reconcilerRedirect) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := s.l.With(
		zap.String("policy", req.NamespacedName),
	)
	res := reconcile.Result{}

	// Fetch the ReplicaSet from the cache
	rs := &balancing.LocalRedirectPolicy{}
	err := s.client.Get(ctx, req.NamespacedName, rs)
	if errors.IsNotFound(err) {
		logger.Sugar().Debugf("policy %s has been deleted", req.NamespacedName)
		// ..... delete ebpf map
		return res, nil
	} else if err != nil {
		logger.Sugar().Debugf("could not fetch %s: %+v", req.NamespacedName, err)
		return res, fmt.Errorf("could not fetch: %+v", err)
	}

	logger.Sugar().Debugf("reconcile: LocalRedirectPolicy policy %s", req.NamespacedName)
	// ........... update the ebpf data

	return res, nil

}
