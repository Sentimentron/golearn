package base

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIDAttribute(t *testing.T) {
	Convey("IDAttribute testing...", t, func() {
		Convey("Set names should be available later...", func() {
			a := &IDAttribute{"hello"}
			So(a.GetName(), ShouldEqual, "hello")
			a.SetName("world")
			So(a.GetName(), ShouldEqual, "world")
			So(a.String(), ShouldEqual, "IDAttribute(world)")
		})
		Convey("ID values should come out unmodified...", func() {
			a := &IDAttribute{""}
			So(a.GetSysValFromString("4"), ShouldResemble, PackU64ToBytes(4))
			So(a.GetSysValFromString("18446744073709551615"), ShouldResemble, PackU64ToBytes(0xFFFFFFFFFFFFFFFF))
			So(a.GetSysValFromString("-1"), ShouldBeNil)
			So(a.GetSysValFromString("a"), ShouldBeNil)
		})
		Convey("Equality should be well defined...", func() {
			a := &IDAttribute{""}
			b := &IDAttribute{""}
			c := NewFloatAttribute("hi")
			So(a.Equals(a), ShouldBeTrue)
			So(a.Equals(b), ShouldBeTrue)
			So(a.Equals(c), ShouldBeFalse)
			a.SetName("hi")
			So(a.Equals(c), ShouldBeFalse)
			So(a.Equals(b), ShouldBeFalse)
			Convey("And so should compatibility...", func() {
				So(a.Compatible(b), ShouldBeTrue)
				So(a.Compatible(c), ShouldBeFalse)
			})
		})
	})
}
