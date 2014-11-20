package base

import (
	"bytes"
	"fmt"
)

type SparseInstances struct {
	d           *DenseInstances        // Holds stuff that must be defined
	c           map[Attribute]bool     // Class Attributes
	s           map[int]map[int][]byte // Sparse storage
	a           map[Attribute]int      // Attribute resolution
	defaultVals map[Attribute][]byte   // defaultValues
	attrCounter int                    // Attribute counter
}

// NewSparseInstances generates a new set of SparseInstances.
// The argument is a slice of class Attributes. New ones can't
// be added at runtime.
func NewSparseInstances(cls []Attribute) *SparseInstances {

	ret := &SparseInstances{
		NewDenseInstances(),
		make(map[Attribute]bool),
		make(map[int]map[int][]byte),
		make(map[Attribute]int),
		make(map[Attribute][]byte),
		0,
	}

	for _, c := range cls {
		ret.d.AddAttribute(c)
		ret.c[c] = true
	}

	return ret

}

// GetAttribute returns an AttributeSpec for a given attribute.
func (s *SparseInstances) GetAttribute(a Attribute) (AttributeSpec, error) {
	// Check in local store
	if v, ok := s.a[a]; ok {
		return AttributeSpec{0, v, a}, nil
	}
	// Returns it from the class pool
	return s.d.GetAttribute(a)

}

// AllAttributes returns all Attributes defined for this SparseInstances.
func (s *SparseInstances) AllAttributes() []Attribute {
	ret := s.d.AllAttributes()
	// Have to sort everything by position
	inv := make([]Attribute, len(s.a))
	for a, i := range s.a {
		inv[i] = a
	}

	for _, a := range inv {
		ret = append(ret, a)
	}
	return ret
}

// AddClassAttribute inserts a class Attribute, as long as Extend() or Set()
// hasn't been called.
func (s *SparseInstances) AddClassAttribute(a Attribute) error {
	// Check that nothing's been allocated yet
	_, rows := s.d.Size()
	if rows > 0 {
		return fmt.Errorf("Can't add class Attribute: already instantiated.")
	}

	// Check that the Attribute is defined...
	_, err := s.GetAttribute(a)
	// If not, return an error
	if err != nil {
		return fmt.Errorf("Class Attribute couldn't be added because it could not be found (error: %s)", err)
	}

	// Set it up as being a class
	s.c[a] = true

	// Add to the underlying DenseInstances
	s.d.AddAttribute(a)
	return s.d.AddClassAttribute(a)
}

// RemoveClassAttribute unsets a given Attribute, as long as Extend() or
// Set() hasn't been called
func (s *SparseInstances) RemoveClassAttribute(a Attribute) error {
	// Remove classhood
	s.c[a] = false

	return s.d.RemoveClassAttribute(a)
}

// AllClassAttributes returns a list of all the defined class Attributes.
func (s *SparseInstances) AllClassAttributes() []Attribute {
	return s.d.AllClassAttributes()
}

// MapOverRows is a convenience function for iteration. Default values
// returned if nothing's explicitly set.
//
// IMPORTANT: rows will not be ordered.
func (s *SparseInstances) MapOverRows(as []AttributeSpec, f func([][]byte, int) (bool, error)) error {

	// Split into class Attributes and not class attributes
	classAttributes := make(map[AttributeSpec]bool)
	nonClassAttributes := make(map[AttributeSpec]bool)

	for _, a := range as {
		if c, ok := s.c[a.GetAttribute()]; ok {
			if c {
				classAttributes[a] = true
				continue
			}
		}
		nonClassAttributes[a] = true
	}
	// Case 1: everything's a class Attribute
	if len(nonClassAttributes) == 0 {
		return s.d.MapOverRows(as, f)
	}
	// Iterate over rows
	buf := make([][]byte, len(as))
	for row := range s.s {
		skipRow := false
		for i, a := range as {
			// If said thing is a class Attribute, call Get on underlying
			var val []byte
			if classAttributes[a] {
				val = s.d.Get(a, row)
			} else {
				val = s.s[row][s.a[a.GetAttribute()]]
			}
			if val == nil || len(val) == 0 {
				// Skip this row
				skipRow = true
			}
			buf[i] = val
		}
		if skipRow {
			continue
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

func (s *SparseInstances) RowString(row int) string {
	var buf bytes.Buffer
	as := ResolveAllAttributes(s)
	for i, a := range as {
		at := a.GetAttribute()
		buf.WriteString(at.GetStringFromSysVal(s.Get(a, i)))
		if i != len(as)-1 {
			buf.WriteString(" ")
		}
	}
	return buf.String()
}

func (s *SparseInstances) Size() (int, int) {
	_, rows := s.d.Size()
	cols := len(s.AllAttributes())
	return cols, rows
}

func (s *SparseInstances) Get(as AttributeSpec, row int) []byte {
	a := as.GetAttribute()
	if s.c[a] {
		// class attribute
		return s.d.Get(as, row)
	}
	// Otherwise, get the position
	p := s.a[a]
	if r, ok := s.s[row]; ok {
		if v, ok := r[p]; ok {
			return v
		}
	}
	return s.defaultVals[a]
}

func (s *SparseInstances) Set(a AttributeSpec, row int, val []byte) {
	pos := a.position
	_, maxRow := s.d.Size()
	if row > maxRow {
		panic("Row out of range! Call Extend() first")
	}
	if s.c[a.GetAttribute()] {
		s.d.Set(a, row, val)
	} else {
		if _, ok := s.s[row]; !ok {
			s.s[row] = make(map[int][]byte)
		}
		s.s[row][pos] = val
	}
}

func (s *SparseInstances) AddAttribute(a Attribute) AttributeSpec {
	var ret AttributeSpec
	ret.position = s.attrCounter
	s.a[a] = s.attrCounter
	s.attrCounter++
	return ret
}

func (s *SparseInstances) Extend(r int) error {
	return s.d.Extend(r)
}

func (s *SparseInstances) SetDefaultValueForAttribute(a Attribute, d interface{}) error {
	val, err := a.GetSysValFromInterface(d)
	if err != nil {
		return err
	}
	s.defaultVals[a] = val
	return nil
}
