package time

import (
	. "github.com/sjwhitworth/golearn/base"
)

// TimeGrid represents datasets accessed in order of time.
type TimeGrid interface {
	FixedDataGrid
	// Gets the value of an Attribute at a given time
	// (can be interpolated)
	At(TimePoint, AttributeSpec) []byte
	// Gets the value of an Attribute at a given time
	// (can't be interpolated)
	PreciselyAt(TimePoint, AttributeSpec) []byte
	// Sets the value of an Attribute at a given time
	Set(TimePoint, []AttributeSpec, [][]byte) error
	// Gets the time Attribute
	GetTimeAttribute() Attribute
}
