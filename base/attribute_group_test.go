package base

import (
	"testing"
)

func TestFloatFloatAttributeGroup(t *testing.T) {
	Convey("Given some rows in iris...", t, func() {
			inst, err := ParseCSVToInstances("../examples/datasets/iris_headers.csv", true)
			So(err, ShouldEqual, nil)
			Convey("Should be able to get the default float column group...", func() {
				fGroup, err := inst.GetDefaultFloatAttributeGroup()
				So(err, ShouldEqual, nil)
				Convey("Size of each row should be 32 bytes", func(){
					So(fGroup.GetRowSize(), ShouldEqual, 32)
				})
				Convey("Rows should return correctly", func() {
					status, _, _, row := getRow(0)
					So(status, ShouldEqual, GetRowSuccess)
					So(len(row), ShouldEqual, 32)
					So(UnpackBytesToFloat(row[0:8], ShouldResemble, 5.1))
					So(UnpackBytesToFloat(row[8:16], ShouldResemble, 3.5))
					So(UnpackBytesToFloat(row[16:24], ShouldResemble, 1.4))
					So(UnpackBytesToFloat(row[24:32], ShouldResemble, 0.2))
				})
			})
	})
}