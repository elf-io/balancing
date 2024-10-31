// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"github.com/elf-io/balancing/pkg/utils" // 替换为你的实际项目路径

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Context("FileExists", func() {
		It("should return true if the file exists", func() {
			Expect(utils.FileExists("utils_test.go")).To(BeTrue())
		})

		It("should return false if the file does not exist", func() {
			Expect(utils.FileExists("non_existent_file.go")).To(BeFalse())
		})
	})

	Context("StringToInt32", func() {
		It("should convert string to int32", func() {
			result, err := utils.StringToInt32("123")
			Expect(err).To(BeNil())
			Expect(result).To(Equal(int32(123)))
		})

		It("should return error for non-numeric string", func() {
			_, err := utils.StringToInt32("abc")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("StringToUint32", func() {
		It("should convert string to uint32", func() {
			result, err := utils.StringToUint32("123")
			Expect(err).To(BeNil())
			Expect(result).To(Equal(uint32(123)))
		})

		It("should return error for non-numeric string", func() {
			_, err := utils.StringToUint32("abc")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("DeepCopy", func() {
		It("should deep copy objects", func() {
			src := map[string]int{"key": 1}
			dst := make(map[string]int)
			err := utils.DeepCopy(&src, &dst)
			Expect(err).To(BeNil())
			Expect(dst).To(Equal(src))
		})
	})
})
