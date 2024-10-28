// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

//go:build !lockdebug
// +build !lockdebug

package lock

import (
	"sync"
)

type internalRWMutex struct {
	sync.RWMutex
}

func (i *internalRWMutex) UnlockIgnoreTime() {
	i.RWMutex.Unlock()
}

type internalMutex struct {
	sync.Mutex
}

func (i *internalMutex) UnlockIgnoreTime() {
	i.Mutex.Unlock()
}
