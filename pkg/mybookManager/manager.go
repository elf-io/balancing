// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package mybookManager

import (
	"github.com/elf-io/elf/pkg/mybookManager/types"
	"go.uber.org/zap"
)

type mybookManager struct {
	logger   *zap.Logger
	webhook  *webhookhander
	informer *informerHandler
}

var _ types.MybookManager = (*mybookManager)(nil)

func New(logger *zap.Logger) types.MybookManager {

	return &mybookManager{
		logger: logger.Named("mybookManager"),
	}
}
