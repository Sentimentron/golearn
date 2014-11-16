package base

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPackedVariableStorageGroup(t *testing.T) {

	Convey("Creating a new test group", t, func() {

		p := NewPackedVariableStorageGroup()
		Convey("Allocating some space...", func() {
			s1 := []byte{0, 1}
			s2 := []byte{83, 12}
			s3 := []byte{1, 1, 2}
			s4 := []byte{1, 7, 3}

			i1, b1 := p.Allocate(len(s1))
			i2, b2 := p.Allocate(len(s2))
			i3, b3 := p.Allocate(len(s3))
			i4, b4 := p.Allocate(len(s4))

			copy(b1, s1)
			copy(b2, s2)
			copy(b3, s3)
			copy(b4, s4)

			Convey("Insert IDs should be right...", func() {
				So(i1, ShouldEqual, 0)
				So(i2, ShouldEqual, 1)
				So(i3, ShouldEqual, 2)
				So(i4, ShouldEqual, 3)
			})

			Convey("Retrieval should work...", func() {
				So(p.Retrieve(0), ShouldResemble, s1)
				So(p.Retrieve(1), ShouldResemble, s2)
				So(p.Retrieve(2), ShouldResemble, s3)
				So(p.Retrieve(3), ShouldResemble, s4)
			})
		})

	})

}
