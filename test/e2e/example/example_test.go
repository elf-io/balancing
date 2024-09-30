// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package example_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("example ", Label("example"), func() {

	It("example", Label("example-1"), func() {

		GinkgoWriter.Printf("deployment  \n")
	})
})
