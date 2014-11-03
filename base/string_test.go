package base

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestStringAttribute(t *testing.T) {
	Convey("StringAttribute testing...", t, func() {
		Convey("Set names should be available later...", func() {
			a := NewStringAttribute("hello")
			So(a.GetName(), ShouldEqual, "hello")
			a.SetName("world")
			So(a.GetName(), ShouldEqual, "world")
			So(a.String(), ShouldEqual, "StringAttribute(world, 0)")
		})
		Convey("ID values should come out unmodified...", func() {
			a := NewStringAttribute("")
			So(a.GetSysValFromString("4"), ShouldResemble, PackU64ToBytes(0))
			So(a.GetSysValFromString("18446744073709551615"), ShouldResemble, PackU64ToBytes(1))
		})
		Convey("Equality should be well defined...", func() {
			a := NewStringAttribute("")
			b := NewStringAttribute("")
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
