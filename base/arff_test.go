package base

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseARFFGetRows(t *testing.T) {
	Convey("Getting the number of rows for a ARFF file", t, func() {
		Convey("With a valid file path", func() {
			numNonHeaderRows := 150
			lineCount, err := ParseARFFGetRows("../examples/datasets/iris.arff")
			So(err, ShouldBeNil)
			So(lineCount, ShouldEqual, numNonHeaderRows)
		})
	})
}

func TestParseARFFGetAttributes(t *testing.T) {
	Convey("Getting the attributes in the headers of a CSV file", t, func() {
		attributes := ParseARFFGetAttributes("../examples/datasets/iris.arff")
		sepalLengthAttribute := attributes[0]
		sepalWidthAttribute := attributes[1]
		petalLengthAttribute := attributes[2]
		petalWidthAttribute := attributes[3]
		speciesAttribute := attributes[4]

		Convey("It gets the correct types for the headers based on the column values", func() {
			_, ok1 := sepalLengthAttribute.(*FloatAttribute)
			_, ok2 := sepalWidthAttribute.(*FloatAttribute)
			_, ok3 := petalLengthAttribute.(*FloatAttribute)
			_, ok4 := petalWidthAttribute.(*FloatAttribute)
			sA, ok5 := speciesAttribute.(*CategoricalAttribute)
			So(ok1, ShouldBeTrue)
			So(ok2, ShouldBeTrue)
			So(ok3, ShouldBeTrue)
			So(ok4, ShouldBeTrue)
			So(ok5, ShouldBeTrue)
			So(sA.GetValues(), ShouldResemble, []string{"iris-setosa", "iris-versicolor", "iris-virginica"})
		})
	})
}

func TestParseARFF(t *testing.T) {
	Convey("Should just be able to load in an ARFF...", t, func() {
		inst, err := ParseDenseARFFToInstances("../examples/datasets/iris.arff")
		So(err, ShouldBeNil)
		So(inst, ShouldNotBeNil)
		So(inst.RowString(0), ShouldEqual, "5.1 3.5 1.4 0.2 Iris-setosa")
		So(inst.RowString(50), ShouldEqual, "7.0 3.2 4.7 1.4 Iris-versicolor")
		So(inst.RowString(100), ShouldEqual, "6.3 3.3 6.0 2.5 Iris-virginica")
	})
}
