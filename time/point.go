package time

import "time"

type timePointMode int

const (
	ArbitraryTime timePointMode = iota
	ConcreteTime
)

// TimePoint represents a specific instant in time, and is
// equivalent to AttributeSpec in base
type TimePoint struct {
	t time.Time // Unexported
	f float64
	m timePointMode
}

func TimePointFromTime(t time.Time) *TimePoint {
	return &TimePoint{
		t,
		0,
		ConcreteTime,
	}
}

func TimePointFromFloat(f float64) *TimePoint {
	return &TimePoint{
		time.Time{},
		f,
		ArbitraryTime,
	}
}
