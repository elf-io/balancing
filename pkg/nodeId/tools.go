package nodeId

import (
	"math/rand"
	"strconv"
	"time"
)

func stringToUint32(str string) (uint32, error) {
	num, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(num), nil
}

func Uint32ToString(num uint32) string {
	return strconv.FormatUint(uint64(num), 10)
}

func generateRandomUint32() uint32 {
	src := rand.New(rand.NewSource(time.Now().UnixNano()))
	return src.Uint32()
}
