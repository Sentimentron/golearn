package time

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestTimePointComparison(t *testing.T) {

	Convey("Testing TimePoint comparisons...", t, func() {

		Convey("With floats...", func() {

			times := make([]TimePoint, 5)
			for i := range times {
				times[i] = TimePointFromFloat(float64(i))
			}

			So(times[0].Compare(times[2]), ShouldEqual, -1)
			So(times[3].Compare(times[3]), ShouldEqual, 0)
			So(times[4].Compare(times[1]), ShouldEqual, 1)

		})

		Convey("With dates...", func() {

			times := make([]TimePoint, 5)
			for i := range times {
				times[i] = TimePointFromTime(time.Date(2014, 8, i, 0, 0, 0, 0, time.UTC))
			}

			So(times[0].Compare(times[2]), ShouldEqual, -1)
			So(times[3].Compare(times[3]), ShouldEqual, 0)
			So(times[4].Compare(times[1]), ShouldEqual, 1)
		})

	})

}
