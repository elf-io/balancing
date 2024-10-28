// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of elf-io

package lock_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lock Suite")
}
