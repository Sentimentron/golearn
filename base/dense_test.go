package base

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestHighDimensionalInstancesLoad(t *testing.T) {
	Convey("Given a high-dimensional dataset...", t, func() {
		_, err := ParseCSVToInstances("../examples/datasets/mnist_train.csv", true)
		So(err, ShouldEqual, nil)
	})
}

func TestHighDimensionalInstancesLoad2(t *testing.T) {
	Convey("Given a high-dimensional dataset...", t, func() {
		// Create the class Attribute
		classAttrs := make(map[int]Attribute)
		classAttrs[0] = NewCategoricalAttribute("Number")
		// Setup the class Attribute to be in its own group
		classAttrGroups := make(map[string]string)
		classAttrGroups["Number"] = "ClassGroup"
		// The rest can go in a default group
		attrGroups := make(map[string]string)

		_, err := ParseCSVToInstancesWithAttributeGroups(
			"../examples/datasets/mnist_train.csv",
			attrGroups,
			classAttrGroups,
			classAttrs,
			true,
		)
		So(err, ShouldEqual, nil)
	})
}

func TestStringAttributeHandling(t *testing.T) {

	Convey("Creating a small sample dataset", t, func() {
		d := NewDenseInstances()
		attrs := make([]Attribute, 2)
		attrs[0] = NewStringAttribute("Hello")
		attrs[1] = NewCategoricalAttribute("World")
		specs := make([]AttributeSpec, 2)
		for i, v := range attrs {
			specs[i] = d.AddAttribute(v)
		}
		d.Extend(2)
		d.Set(specs[0], 0, attrs[0].GetSysValFromString("Hello World"))
		d.Set(specs[1], 0, attrs[1].GetSysValFromString("Greeting"))
		d.Set(specs[0], 1, attrs[0].GetSysValFromString("Goodbye cruel world!"))
		d.Set(specs[0], 1, attrs[1].GetSysValFromString("Farewell"))
		Convey("Rows should appear correct...", func() {
			So(d.RowString(0), ShouldEqual, "Hello World   Greeting")
			So(d.RowString(1), ShouldEqual, "Goodbye cruel world   Farewell")
		})
	})

}
