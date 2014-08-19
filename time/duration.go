package time

import "time"

type TimeDuration struct {
	d time.Duration
	f float64
	m timePointMode
}

func TimeDurationFromDuration(d time.Duration) TimeDuration {
	return TimeDuration{
		d,
		0,
		ConcreteTime,
	}
}

func TimeDurationFromFloat(f float64) TimeDuration {
	return TimeDuration{
		0,
		f,
		ArbitraryTime,
	}
}

func (t TimeDuration) Compare(other TimeDuration) int {
	if t.m != other.m {
		panic("Invalid comparison")
	} else if t.m == ArbitraryTime {
		if t.f > other.f {
			return 1
		} else if t.f == other.f {
			return 0
		} else {
			return -1
		}
	} else if t.m == ConcreteTime {
		if t.d > other.d {
			return 1
		} else if t.d == other.d {
			return 0
		} else {
			return -1
		}
	} else {
		panic("Invalid comparison!")
	}
}
