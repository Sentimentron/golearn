package base

import (
	"database/sql"
)

func DenseInstancesFromSQLRows(rows *sql.Rows, attrs []Attribute) (*DenseInstances, error) {

	// Create return structure
	d := NewDenseInstances()
	specs := make([]AttributeSpec, len(attrs))
	for i, a := range attrs {
		specs[i] = d.AddAttribute(a)
	}

	// Create temporary column buffer
	buf := make([][]string, 0)
	for rows.Next() {
		tmp := make([]string, len(attrs))
		tmp2 := make([]interface{}, len(attrs))
		for i := range tmp2 {
			tmp2[i] = &tmp[i]
		}
		err := rows.Scan(tmp2...)
		if err != nil {
			return nil, err
		}
		buf = append(buf, tmp)
	}

	// Allocate instances memory
	d.Extend(len(buf))
	for i, b := range buf {
		for j, v := range b {
			val := attrs[j].GetSysValFromString(v)
			d.Set(specs[j], i, val)
		}
	}

	return d, nil
}
