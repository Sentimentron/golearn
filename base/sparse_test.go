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

func TestSparseDefaultAttributeHandling(t *testing.T) {
	Convey("Creating a small sample dataset", t, func() {
		d := NewSparseInstances()
		attrs := make([]Attribute, 2)
		attrs[0] = NewStringAttribute("Hello")
		attrs[1] = NewCategoricalAttribute("World")
		specs := make([]AttributeSpec, 2)
		for i, v := range attrs {
			specs[i] = d.AddAttribute(v)
		}

		err := d.SetDefaultValueForAttribute(attrs[0], "Whatever")
		So(err, ShouldBeNil)

		err = d.AddClassAttribute(attrs[1])
		So(err, ShouldBeNil)

		// Check initial size
		cols, rows := d.Size()
		So(cols, ShouldEqual, 2)
		So(rows, ShouldEqual, 0)

		d.Set(specs[0], 0, attrs[0].GetSysValFromString("Hello World"))
		d.Set(specs[1], 0, attrs[1].GetSysValFromString("Greeting"))

		// Check next size
		cols, rows = d.Size()
		So(cols, ShouldEqual, 2)
		So(rows, ShouldEqual, 1)

		d.Set(specs[1], 1, attrs[1].GetSysValFromString("Farewell"))

		// Check final size
		cols, rows = d.Size()
		So(cols, ShouldEqual, 2)
		So(rows, ShouldEqual, 2)

		Convey("Rows should appear correct...", func() {
			So(d.RowString(0), ShouldEqual, "Hello World Greeting")
			So(d.RowString(1), ShouldEqual, "Whatever Farewell")
		})

		Convey("Data should be correct in MapOverRows...", func() {
			rows := 0
			err := d.MapOverRows(specs, func(v [][]byte, r int) (bool, error) {
				if r == 0 {
					So(attrs[0].GetStringFromSysVal(v[0]), ShouldEqual, "Hello World")
					So(attrs[1].GetStringFromSysVal(v[1]), ShouldEqual, "Greeting")
				} else if r == 1 {
					So(attrs[0].GetStringFromSysVal(v[0]), ShouldEqual, "Whatever")
					So(attrs[1].GetStringFromSysVal(v[1]), ShouldEqual, "Farewell")
				}
				So(r, ShouldBeLessThan, 2)
				rows++
				return true, nil
			})
			So(err, ShouldBeNil)
			So(rows, ShouldEqual, 2)
		})
	})
}
