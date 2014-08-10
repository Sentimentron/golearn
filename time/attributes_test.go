package time

import (
	"github.com/sjwhitworth/golearn/base"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestEpochNSTimeAttribute(t *testing.T) {
	Convey("Testing some abstract properties of EpochNSTimeAttribute", t, func() {
		a := NewEpochNSTimeAttribute("Hello", time.RFC3339)
		Convey("Checking GetName() and SetName()", func() {
			So(a.GetName(), ShouldEqual, "Hello")
			a.SetName("World")
			So(a.GetName(), ShouldEqual, "World")
		})
		Convey("Checking parse from date...", func() {
			val := a.GetSysValFromString("2014-07-24T12:16:36Z")
			So(base.UnpackBytesToU64(val), ShouldEqual, 1406204196)
		})
		Convey("Checking conversion to date...", func() {
			val := a.GetStringFromSysVal(base.PackU64ToBytes(1406204196))
			So(val, ShouldEqual, "2014-07-24T12:16:36Z")
		})
	})
}
