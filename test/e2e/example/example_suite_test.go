// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package example_test

import (
	"testing"

	e2e "github.com/elf-io/e2eframework/framework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	// "k8s.io/apimachinery/pkg/runtime"
)

func TestAssignIP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "example Suite")
}

var frame *e2e.Framework

var _ = BeforeSuite(func() {
	defer GinkgoRecover()
	var e error
	frame, e = e2e.NewFramework(GinkgoT(), nil)
	Expect(e).NotTo(HaveOccurred())

})
