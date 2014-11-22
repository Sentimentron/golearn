package base

import (
	"bytes"
	"fmt"
)

// SparseInstances are used to represent sparse data.
type SparseInstances struct {
	c           map[Attribute]bool     // Class Attributes
	s           map[int]map[int][]byte // Sparse storage
	a           map[Attribute]int      // Attribute resolution
	defaultVals map[int][]byte         // defaultValues
	attrCounter int                    // Attribute counter
	maxRow      int
}

// NewSparseInstances generates a new set of SparseInstances.
func NewSparseInstances() *SparseInstances {
	return &SparseInstances{
		make(map[Attribute]bool),
		make(map[int]map[int][]byte),
		make(map[Attribute]int),
		make(map[int][]byte),
		0,
		0,
	}
}

// GetAttribute returns an AttributeSpec for a given attribute.
func (s *SparseInstances) GetAttribute(a Attribute) (AttributeSpec, error) {
	// Check in local store
	if v, ok := s.a[a]; ok {
		return AttributeSpec{0, v, a}, nil
	}
	return AttributeSpec{0, 0, nil}, fmt.Errorf("Could not resolve Attribute %s", a)

}

// AllAttributes returns all Attributes defined for this SparseInstances.
func (s *SparseInstances) AllAttributes() []Attribute {

	// Have to sort everything by position
	inv := make([]Attribute, len(s.a))
	for a, i := range s.a {
		inv[i] = a
	}

	return inv
}

// AddClassAttribute inserts a class Attribute, as long as Extend() or Set()
// hasn't been called.
func (s *SparseInstances) AddClassAttribute(a Attribute) error {

	// Check that the Attribute is defined...
	_, err := s.GetAttribute(a)
	// If not, return an error
	if err != nil {
		return fmt.Errorf("Class Attribute couldn't be added because it could not be found (error: %s)", err)
	}

	// Set it up as being a class
	s.c[a] = true
	return nil
}

// RemoveClassAttribute unsets a given Attribute, as long as Extend() or
// Set() hasn't been called
func (s *SparseInstances) RemoveClassAttribute(a Attribute) error {
	// Remove classhood
	s.c[a] = false
	return nil
}

// AllClassAttributes returns a list of all the defined class Attributes.
func (s *SparseInstances) AllClassAttributes() []Attribute {
	var ret []Attribute
	for a := range s.c {
		if s.c[a] {
			ret = append(ret, a)
		}
	}
	return ret
}

// MapOverRows is a convenience function for iteration. Default values
// returned if nothing's explicitly set. If the default value is missing
// or set to nil, the entire row's skipped.
//
// IMPORTANT: rows will not be ordered.
func (s *SparseInstances) MapOverRows(as []AttributeSpec, f func([][]byte, int) (bool, error)) error {
	// Iterate over rows
	buf := make([][]byte, len(as))
	for row := range s.s {
		if _, ok := s.s[row]; !ok {
			continue
		}
		for i, a := range as {
			val := s.s[row][a.position]
			if val == nil || len(val) == 0 {
				val = s.defaultVals[a.position]
			}
			if val == nil {
				panic(fmt.Errorf("Nil defined value: %s on line %d", a.GetAttribute(), row))
			}
			buf[i] = val
		}

		// Call the user defined function
		next, err := f(buf, row)
		if err != nil {
			return err
		}
		if !next {
			break
		}
	}
	return nil
}

// RowString returns a string representing the values on a given
// line.
func (s *SparseInstances) RowString(row int) string {
	var buf bytes.Buffer
	as := ResolveAllAttributes(s)
	for i, a := range as {
		at := a.GetAttribute()
		buf.WriteString(at.GetStringFromSysVal(s.Get(a, row)))
		if i != len(as)-1 {
			buf.WriteString(" ")
		}
	}
	return buf.String()
}

// Size returns the dimensions of this SparseInstances.
// First value is the number of columns, second is the number of rows.
func (s *SparseInstances) Size() (int, int) {
	cols := len(s.AllAttributes())
	return cols, s.maxRow
}

// Get retrieves the []byte slice stored at a given AttributeSpec, row
// coordinate.
func (s *SparseInstances) Get(as AttributeSpec, row int) []byte {
	if r, ok := s.s[row]; ok {
		if v, ok := r[as.position]; ok {
			return v
		}
	}
	if _, ok := s.defaultVals[as.position]; !ok {
		panic(fmt.Errorf("No default value set for %s", as.GetAttribute()))
	}
	return s.defaultVals[as.position]
}

// Set sets the []byte slice at a given AttributeSpec, row coordinate.
func (s *SparseInstances) Set(a AttributeSpec, row int, val []byte) {
	pos := a.position
	if _, ok := s.s[row]; !ok {
		s.s[row] = make(map[int][]byte)
	}
	s.s[row][pos] = val
	if row >= s.maxRow {
		s.maxRow = row + 1
	}
}

// AddAttribute inserts an Attribute and then returns a specification.
func (s *SparseInstances) AddAttribute(a Attribute) AttributeSpec {
	var ret AttributeSpec
	ret.position = s.attrCounter
	s.a[a] = s.attrCounter
	s.attrCounter++
	return ret
}

// Extend increases the number of defined rows.
func (s *SparseInstances) Extend(r int) error {
	s.maxRow += r
	return nil
}

// SetDefaultValueForAttribute sets what value is returned for a given
// Attribute if nothing's been set.
func (s *SparseInstances) SetDefaultValueForAttribute(a Attribute, d interface{}) error {
	val, err := a.GetSysValFromInterface(d)
	if err != nil {
		return err
	}

	as, err := s.GetAttribute(a)
	if err != nil {
		return err
	}

	s.defaultVals[as.position] = val
	return nil
}
