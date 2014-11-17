package base

import (
	"bytes"
	"fmt"
	"sync"
)

// This type can be used as the backing storage for VariableAttributes.
type PackedVariableStorageGroup struct {
	m sync.Mutex // Sync
	v [][]byte   // Storage
}

func NewPackedVariableStorageGroup() *PackedVariableStorageGroup {
	return &PackedVariableStorageGroup{
		sync.Mutex{},
		make([][]byte, 0),
	}
}

// Allocate allocates and/or returns size bytes for permanent storage.
func (p *PackedVariableStorageGroup) Allocate(size int) (uint64, []byte) {
	p.m.Lock()
	defer p.m.Unlock()

	ret := make([]byte, size)
	p.v = append(p.v, ret)
	return uint64(len(p.v) - 1), ret
}

// Returns a given set of bytes
func (p *PackedVariableStorageGroup) Retrieve(off uint64) []byte {
	return p.v[off]
}

// VariableAttributeGroups store columns with variable lengths.
type VariableAttributeGroup struct {
	parent     DataGrid
	attributes []VariableAttribute
	size       int
	alloc      []byte
	maxRow     int
	p          *PackedVariableStorageGroup
}

func (v *VariableAttributeGroup) String() string {
	return "VariableAttributeGroup"
}

func (v *VariableAttributeGroup) RowSizeInBytes() int {
	return 8 * len(v.attributes)
}

func (v *VariableAttributeGroup) Attributes() []Attribute {
	ret := make([]Attribute, len(v.attributes))
	for i, a := range v.attributes {
		ret[i] = a
	}
	return ret
}

func (v *VariableAttributeGroup) AddAttribute(a Attribute) error {
	if attr, ok := a.(VariableAttribute); ok {
		v.attributes = append(v.attributes, attr)
		return nil
	}
	return fmt.Errorf("%s is not a VariableAttribute!", a)
}

func (v *VariableAttributeGroup) setStorage(a []byte) {
	v.alloc = a
}

func (v *VariableAttributeGroup) Storage() []byte {
	return v.alloc
}

func (v *VariableAttributeGroup) offset(col, row int) int {
	return row*v.RowSizeInBytes() + col*8
}

func (v *VariableAttributeGroup) set(col, row int, val []byte) {
	// Find the attribute
	a := v.attributes[col]
	// Get the length
	l := a.GetLengthFromSysVal(val)
	// Allocate space in the storage group
	i, al := v.p.Allocate(l)
	copy(al, val)
	// Figure out how to set
	str := PackU64ToBytes(i)
	// Find the right place
	offset := v.offset(col, row)
	copied := copy(v.alloc[offset:], str)
	if copied != 8 {
		panic(fmt.Sprintf("set() terminated with only copying %d bytes", copied))
	}
	row++
	if row > v.maxRow {
		v.maxRow = row
	}
}

func (v *VariableAttributeGroup) get(col, row int) []byte {
	offset := v.offset(col, row)
	str := v.alloc[offset : offset+8]
	off := UnpackBytesToU64(str)
	return v.p.Retrieve(off)
}

func (v *VariableAttributeGroup) appendToRowBuf(row int, buffer *bytes.Buffer) {
	for i, a := range v.attributes {
		postfix := " "
		if i == len(v.attributes)-1 {
			postfix = ""
		}
		buffer.WriteString(fmt.Sprintf("%s%s", a.GetStringFromSysVal(v.get(i, row)), postfix))

	}
}

func (v *VariableAttributeGroup) resize(add int) {
	newAlloc := make([]byte, len(v.alloc)+add)
	copy(newAlloc, v.alloc)
	v.alloc = newAlloc
}
