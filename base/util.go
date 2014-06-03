package base

import (
	"encoding/binary"
	"fmt"
	"math"
)

func PackU64ToBytes(val uint64) []byte {
	ret := make([]byte, 8)
	status := binary.PutUvarint(ret, val)
	if status != 1 {
		panic(fmt.Sprintf("Packing failed! %d", status))
	}
	return ret
}

func UnpackBytesToU64(val []byte) uint64 {
	intVal, status := binary.Uvarint(val)
	if status != 1 {
		panic(fmt.Sprintf("Unpacking failed! %d", status))
	}
	return intVal
}

func PackFloatToBytes(val float64) []byte {
	return PackU64ToBytes(math.Float64bits(val))
}

func UnpackBytesToFloat(val []byte) float64 {
	return math.Float64FromBits(UnpackBytesToU64(val))
}
