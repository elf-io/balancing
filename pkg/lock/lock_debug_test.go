// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of elf-io

//go:build lockdebug
// +build lockdebug

package lock_test

import (
	"github.com/elf-io/balancing/pkg/lock"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("LockFast", Label("unitest"), func() {

	// it is daemon , add more test here
	It("test debug lock", func() {

		l := &lock.Mutex{}

		l.Lock()
		l.Unlock()
	})
	It("test debug Rlock", func() {
		l := &lock.RWMutex{}
		l.RLock()
		l.RUnlock()
		l.Lock()
		l.Unlock()
	})
})
