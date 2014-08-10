package time

import (
	. "github.com/sjwhitworth/golearn/base"
	"time"
)

// IrregularTimeSeries represents time-ordered instances
// where measurements occur without a fixed duration.
type IrregularTimeSeries struct {
	DenseInstances
	points       []TimePoint
	timeAttrSpec AttributeSpec
	nextRow      int
}

// NewIrregularTimeSeries creates a new IrregularTimeSeries.
// By default, this includes a time Attribute in RFC3339.
func NewIrregularTimeSeries() *IrregularTimeSeries {
	// Create the time series
	ret := &IrregularTimeSeries{
		*NewDenseInstances(),
		make([]TimePoint, 0),
		AttributeSpec{},
		0,
	}
	// Create the time Attribute
	attr := NewEpochNSTimeAttribute("Time", time.RFC3339)
	ret.timeAttrSpec = ret.AddAttribute(attr)
	return ret
}

// Returns where the tp TimePoint should be in this IrregularTimeSeries.
// Need to check for equivalence with this point.
func (t *IrregularTimeSeries) searchForTimePointOffset(tp TimePoint) int {
	low := 0
	high := len(t.points)
	for {
		if low >= high {
			return low
		}
		test := (low + high) / 2
		comp := t.points[test].Compare(tp)
		if comp == 0 {
			return test
		} else if comp == -1 { // test point's less than me
			low = test + 1 // Final point must be above that
		} else {
			high = test // Ultimate boundary is higher than here
		}
	}
}

func (t *IrregularTimeSeries) insertTimePointAtOffset(tp TimePoint, offset int) {
	shift := t.points[offset+1:]
	stay := t.points[:offset+1]
	t.points = append(stay, tp)
	t.points = append(t.points, shift...)
}

func (t *IrregularTimeSeries) lookupRow(tp TimePoint) (int, bool) {
	offset := t.searchForTimePointOffset(tp)
	tRef := t.points[offset]
	if tRef.Compare(tp) == 0 {
		return offset, true
	}
	return offset, false
}

// AddAttribute adds an Attribute to this IrregularTimeSeries.
// panics() if the Attribute is named the same as the time Attribute.
func (t *IrregularTimeSeries) AddAttribute(a Attribute) AttributeSpec {
	if a.GetName() == "Time" {
		if len(t.AllAttributes()) != 0 {
			panic("Attribute cannot be called Time!")
		}
	}
	return t.DenseInstances.AddAttribute(a)
}

// PreciselyAt returns byte sequence for a given AttributeSpec at
// a precise timepoint, and nil if no observations are recorded.
func (t *IrregularTimeSeries) PreciselyAt(tp TimePoint, a AttributeSpec) []byte {
	// Lookup the row
	if row, ok := t.lookupRow(tp); !ok {
		return nil
	} else {
		return t.DenseInstances.Get(a, row)
	}
}

// At returns a byte sequence for a given AttributeSpec at any TimePoint
//
// If the AttributeSpec describes a FloatAttribute, the result can be interpolated.
// Otherwise, the value returned will be that of the TimePoint which is closest.
func (t *IrregularTimeSeries) At(tp TimePoint, a AttributeSpec) []byte {

	// Look for an exact match
	offset, ok := t.lookupRow(tp)
	if ok {
		return t.DenseInstances.Get(a, offset)
	}

	return nil
}

// Set records a sequence of observations at a given time.
func (t *IrregularTimeSeries) Set(tp TimePoint, as []AttributeSpec, b [][]byte) error {

	// If we can't find the row, insert it
	if offset, ok := t.lookupRow(tp); !ok {
		t.insertTimePointAtOffset(tp, offset)
		t.nextRow++
	}

	// Get the row
	row, _ := t.lookupRow(tp)

	// Set each underlying AttributeSpec
	for i, a := range as {
		t.DenseInstances.Set(a, row, b[i])
	}

	return nil
}
