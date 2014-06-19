package base

type AddAttributeStatus int
const (
	AddAttributeWrongType AddAttributeStatus = iota
	AddAttributeNotAllowed
	AddAttributeSuccess
)

type addMemoryStatus int
const (
	AddedMemorySuccess  addMemoryStatus = iota
	AddedMemoryWrongSize
)

type getRowStatus int
const (
	GetRowRangeError getRowStatus = iota
	GetRowNextSliceError
	GetRowSuccess
)

type AttributeGroup interface {
	// Returns the physical size of each row in this AttributeGroup
	GetRowSize() uint

	// Add a column to this AttributeGroup
	AddAttribute(Attribute) AddAttributeStatus

	// Checks whether this Attribute is in here
	HasAttribute(Attribute) bool

	// Add additional memory to the control of this Attribute
	addMemory([]byte) addMemoryStatus

	// Gets the raw bytes representing a row
	// First int return argument is the memory reference
	// Second int is the row offset in that reference
	getRow(uint) (getRowStatus, uint, uint, []byte)

	// Gets the raw bytes representing a row
	// Re-entrant call from a previous getRow call
	// First int is the memory reference
	// Second int is the desired row offset (relative to the start)
	// First return int is the memory reference
	getRowIter(uint, uint) (getRowStatus, []byte)
}

type FloatAttributeGroup struct {
	// Stores the number of Attributes represented in this group
	attrCount uint
	// Stores the total number of rows we have allocated
	rowTotal  uint
	// Maps each Attribute to a offset (physical offset is byteLen * this number)
	attrMapping map[Attribute] uint
	// Stores references to the byte slies managed by this FloatAttributeGroup
	memorySlices [][]byte
	// Stores the number of rows inside each memory slice
	memoryRows   []uint
}

func (f *FloatAttributeGroup) GetRowSize() uint {
	return f.attrCount * 8
}

func (f *FloatAttributeGroup) AddAttribute(srcA Attribute) AddAttributeStatus {
	if len(f.memory) > 0 {
		return AddAttributeNotAllowed
	}
	if a, ok = srcA.(*base.FloatAttribute); !ok {
		return AddAttributeWrongType
	} else {
		f.attrMapping[srcA] = f.attrCount
		f.attrCount++
		return AddAttributeSuccess
	}
}

func (f *FloatAttributeGroup) HasAttribute(a Attribute) bool {
	_, ok := f.attrMapping[a]
	return ok
}

func (f *FloatAttributeGroup) addMemory(b []byte) addMemoryStatus {
	rowSize := f.GetRowSize()
	memorySize := len(b)
	rowAlloc := memorySize / rowSize
	if rowAlloc * rowSize > memorySize {
		rowAlloc--
	}
	if rowAlloc == 0 {
		return AddedMemoryWrongSize
	}
	f.memorySlices = append(f.memorySlices, b)
	f.memoryRows = append(f.memoryRows, rowAlloc)
	f.rowTotal += rowAlloc
	return AddedMemorySuccess
}

func (f *FloatAttributeGroup) getRow(row uint) (GetRowStatus, uint, uint, []byte) {
	var rowSum uint
	var slice uint
	var rowAlloc uint

	// Double check we have that row
	if row > f.rowAlloc {
		return GetRowRangeError, 0, nil
	}

	// Find the slice which contains the row
	for slice, rowAlloc := range f.memoryRows {
		rowSum += rowAlloc
		if rowSum > row {
			slice--
			break
		}
	}

	// Compute the row offset
	rowOffset := row % rowAlloc

	// Compute the row size
	rowSize := f.GetRowSize()

	// Get the row slice
	physicalOffset := rowSize * rowOffset
	ret := f.memorySlices[slice][physicalOffset : physicalOffset + rowSize + 1]
	return GetRowSuccess, slice, rowOffset, ret
}

func (f *FloatAttributeGroup) getRowIter(rowOffset uint, slice uint) (getRowStatus, []byte) {
	// Check if the slice has that row
	rowAlloc := f.memoryRows[slice]
	if rowAlloc < rowOffset {
		slice += 1
		if slice >= len(f.memoryRows) {
			// Out of bounds on the last slice
			return GetRowRangeError, 0, nil
		}
		return GetRowNextSliceError, slice, nil
	}
	// Compute the row size
	rowSize := f.GetRowSize()

	// Get the row slice
	physicalOffset := rowSize * rowOffset
	ret := f.memorySlices[slice][physicalOffset : physicalOffset + rowSize + 1]
	return GetRowSuccess, slice, ret
}
