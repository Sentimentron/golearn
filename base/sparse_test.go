package base

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSparseStringAttributeHandling(t *testing.T) {

	Convey("Creating a small sample dataset", t, func() {
		d := NewSparseInstances()
		attrs := make([]Attribute, 2)
		attrs[0] = NewStringAttribute("Hello")
		attrs[1] = NewCategoricalAttribute("World")
		specs := make([]AttributeSpec, 2)
		for i, v := range attrs {
			specs[i] = d.AddAttribute(v)
		}
		err := d.AddClassAttribute(attrs[1])
		So(err, ShouldBeNil)
		d.Extend(2)
		d.Set(specs[0], 0, attrs[0].GetSysValFromString("Hello World"))
		d.Set(specs[1], 0, attrs[1].GetSysValFromString("Greeting"))
		d.Set(specs[0], 1, attrs[0].GetSysValFromString("Goodbye cruel world!"))
		d.Set(specs[1], 1, attrs[1].GetSysValFromString("Farewell"))
		Convey("Rows should appear correct...", func() {
			So(d.RowString(0), ShouldEqual, "Hello World Greeting")
			So(d.RowString(1), ShouldEqual, "Goodbye cruel world! Farewell")
		})
	})

}
