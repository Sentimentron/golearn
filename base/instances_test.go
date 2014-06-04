package base

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestInstancesAppendRowMinimal(t *testing.T) {
	attr := NewFloatAttribute()
	attr.SetName("x")
	attrs := make([]Attribute, 1)
	attrs[0] = attr
	inst := NewInstances(attrs, 1)
	Convey("Given some float", t, func() {
		x := "1.21"
		Convey("When the float gets added as a row", func() {
			packed := attr.GetSysValFromString(x)
			rowMap := make(map[Attribute][]byte)
			rowMap[attr] = packed
			err := inst.AppendRow(rowMap)
			Convey("That shouldn't have an error", func() {
				So(err, ShouldEqual, nil)
			})
			fmt.Println(inst.storage)
			Convey("And the row gets printed", func() {
				rowStr := inst.RowStr(0)
				Convey("The value in the row string should be the same", func() {
					So(rowStr, ShouldEqual, x)
				})
			})
		})
	})
}
