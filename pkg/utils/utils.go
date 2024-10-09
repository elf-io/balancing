// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
