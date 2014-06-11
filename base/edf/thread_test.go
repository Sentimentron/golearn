package edf

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestThreadDeserialize(T *testing.T) {
	bytes := []byte{0, 0, 0, 6, 83, 89, 83, 84, 69, 77, 0, 0, 0, 1}
	Convey("Given a byte slice", T, func() {
		var t Thread
		size := t.Deserialize(bytes)
		Convey("Decoded name should be SYSTEM", func() {
			So(t.name, ShouldEqual, "SYSTEM")
		})
		Convey("Size should be the same as the array", func() {
			So(size, ShouldEqual, len(bytes))
		})
	})
}

func TestThreadSerialize(T *testing.T) {
	var t Thread
	refBytes := []byte{0, 0, 0, 6, 83, 89, 83, 84, 69, 77, 0, 0, 0, 1}
	t.name = "SYSTEM"
	t.id = 1
	toBytes := make([]byte, len(refBytes))
	Convey("Should serialize correctly", T, func() {
		t.Serialize(toBytes)
		So(toBytes, ShouldResemble, refBytes)
	})
}
