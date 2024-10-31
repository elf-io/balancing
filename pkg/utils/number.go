package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

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

// RandomUint32 returns a randomly generated uint32 number.
func RandomUint32() uint32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Uint32()
}
