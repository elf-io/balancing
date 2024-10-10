// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"os"
	"strconv"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

const (
	int32Min = -1 << 31
	int32Max = 1<<31 - 1
)

func StringToInt32(s string) (int32, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if i < int32Min || i > int32Max {
		return 0, fmt.Errorf("value out of int32 range: %d", i)
	}
	return int32(i), nil
}
