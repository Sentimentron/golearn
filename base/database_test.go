package base

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDenseInstancesFromSQLRows(t *testing.T) {

	Convey("Opening the iris.sqlite file...", t, func() {
		db, err := sql.Open("sqlite3", "../examples/datasets/iris.sqlite")
		So(err, ShouldBeNil)
		Convey("Selecting all the data...", func() {
			rows, err := db.Query("SELECT * FROM iris")
			defer rows.Close()
			So(err, ShouldBeNil)
			Convey("Creating the instances...", func() {
				attrs := make([]Attribute, 5)
				attrs[0] = NewFloatAttribute("Sepal Length")
				attrs[1] = NewFloatAttribute("Sepal Width")
				attrs[2] = NewFloatAttribute("Petal Length")
				attrs[3] = NewFloatAttribute("Petal Width")
				attrs[4] = NewCategoricalAttribute("Species")
				d, err := DenseInstancesFromSQLRows(rows, attrs)
				So(err, ShouldBeNil)
				cols, rows := d.Size()
				So(rows, ShouldEqual, 150)
				So(cols, ShouldEqual, 5)
				So(d.RowString(0), ShouldEqual, "5.10 3.50 1.40 0.20 Iris-setosa")
				So(d.RowString(50), ShouldEqual, "7.00 3.20 4.70 1.40 Iris-versicolor")
				So(d.RowString(100), ShouldEqual, "6.30 3.30 6.00 2.50 Iris-virginica")
				So(d.RowString(149), ShouldEqual, "5.90 3.00 5.10 1.80 Iris-virginica")
			})
		})
	})

}
