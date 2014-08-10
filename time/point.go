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

func TimePointFromTime(t time.Time) TimePoint {
	return TimePoint{
		t,
		0,
		ConcreteTime,
	}
}

func TimePointFromFloat(f float64) TimePoint {
	return TimePoint{
		time.Time{},
		f,
		ArbitraryTime,
	}
}

func (t TimePoint) Compare(other TimePoint) int {
	// Double check
	if t.m != other.m {
		panic("Wrong comparison")
	} else if t.m == ArbitraryTime {
		if t.f > other.f {
			return 1
		} else if t.f == other.f {
			return 0
		} else {
			return -1
		}
	} else if t.m == ConcreteTime {
		u := t.t.Unix()
		v := other.t.Unix()
		if u > v {
			return 1
		} else if u == v {
			return 0
		} else {
			return -1
		}
	} else {
		panic("Invalid comparison!")
	}
}
