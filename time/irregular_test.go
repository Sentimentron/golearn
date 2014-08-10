package time

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIrregularTimeSeriesListOps(t *testing.T) {

	Convey("Generating 5 time-stamps", t, func() {
		times := make([]TimePoint, 5)
		for i := range times {
			times[i] = TimePointFromFloat(float64(i))
		}

		ir := NewIrregularTimeSeries()
		ir.points = times
		Convey("Should be able to find things...", func() {
			So(ir.searchForTimePointOffset(times[0]), ShouldEqual, 0)
			So(ir.searchForTimePointOffset(times[1]), ShouldEqual, 1)
			So(ir.searchForTimePointOffset(times[2]), ShouldEqual, 2)
			So(ir.searchForTimePointOffset(times[3]), ShouldEqual, 3)
			So(ir.searchForTimePointOffset(times[4]), ShouldEqual, 4)

			Convey("Insertion should occur correctly...", func() {
				Convey("At the bottom...", func() {
					ir.points = append(make([]TimePoint, 0), times...)
					t := TimePointFromFloat(6.0)
					ir.insertTimePointAtOffset(t, 0)
					So(ir.points[0].Compare(t), ShouldEqual, 0)
					for i, ts := range times[:4] {
						fmt.Println(i + 1)
						So(ir.points[i+1].Compare(ts), ShouldEqual, 0)
					}

					Convey("At the top...", func() {
						ir.points = append(make([]TimePoint, 0), times...)
						t := TimePointFromFloat(6.0)
						ir.insertTimePointAtOffset(t, 4)
						for i, ts := range times[:4] {
							So(ir.points[i].Compare(ts), ShouldEqual, 0)
						}
						So(ir.points[4].Compare(t), ShouldEqual, 0)

						Convey("In the middle...", func() {
							ir.points = append(make([]TimePoint, 0), times...)
							t := TimePointFromFloat(6)
							ir.insertTimePointAtOffset(t, 2)
						})
					})
				})
			})

		})

	})

}
