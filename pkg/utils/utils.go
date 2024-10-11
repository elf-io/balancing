// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"encoding/gob"
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

func DeepCopy(src, dst interface{}) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	if err := enc.Encode(src); err != nil {
		return err
	}
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func StringToUint32(str string) (uint32, error) {
	if str == "" {
		return 0, fmt.Errorf("empty string")
	}
	num, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, err
	}
	if num > uint64(uint32(^uint32(0))) {
		return 0, fmt.Errorf("exceed the uint32")
	}
	return uint32(num), nil
}
