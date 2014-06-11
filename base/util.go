package base

import (
	"math"
)

func PackU64ToBytes(val uint64) []byte {
	ret := make([]byte, 8)
	ret[0] = byte(val & (0xFF << 56) >> 56)
	ret[1] = byte(val & (0xFF << 48) >> 48)
	ret[2] = byte(val & (0xFF << 40) >> 40)
	ret[3] = byte(val & (0xFF << 32) >> 32)
	ret[4] = byte(val & (0xFF << 24) >> 24)
	ret[5] = byte(val & (0xFF << 16) >> 16)
	ret[6] = byte(val & (0xFF << 8) >> 8)
	ret[7] = byte(val & (0xFF << 0) >> 0)
	return ret
}

func UnpackBytesToU64(val []byte) uint64 {
	intVal := uint64(0)
	intVal |= uint64(val[0]) << 56
	intVal |= uint64(val[1]) << 48
	intVal |= uint64(val[2]) << 40
	intVal |= uint64(val[3]) << 32
	intVal |= uint64(val[4]) << 24
	intVal |= uint64(val[5]) << 16
	intVal |= uint64(val[6]) << 8
	intVal |= uint64(val[7]) << 0
	return intVal
}

func PackFloatToBytes(val float64) []byte {
	return PackU64ToBytes(math.Float64bits(val))
}

func UnpackBytesToFloat(val []byte) float64 {
	return math.Float64frombits(UnpackBytesToU64(val))
}

func GeneratePredictionVector(from FixedDataGrid) FixedDataGrid {
	classAttrsMap := from.GetClassAttrs()
	classAttrs := make([]Attribute, 0)
	for attr := range classAttrsMap {
		classAttrs = append(classAttrs, classAttrsMap[attr])
	}
	_, rowCount := from.Size()
	ret := NewInstances(classAttrs, int(rowCount))
	return ret
}
