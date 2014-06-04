package base

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestFloatAttributeSysVal(t *testing.T) {
	Convey("Given some float", t, func() {
		x := "1.21"
		attr := NewFloatAttribute()
		Convey("When the float gets packed", func() {
			packed := attr.GetSysValFromString(x)
			Convey("And then unpacked", func() {
				unpacked := attr.GetStringFromSysVal(packed)
				Convey("The unpacked version should be the same", func() {
					So(unpacked, ShouldEqual, x)
				})
			})
		})
	})
}
