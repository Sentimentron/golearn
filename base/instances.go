package base

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sjwhitworth/golearn/base/edf"
	"math/rand"
)

// Instances represents a grid of numbers (typed by Attributes)
// stored internally in mat.DenseMatrix as float64's.
// See docs/instances.md for more information.
type Instances struct {
	storage    *edf.EdfFile
	attributes []Attribute
	attrLookup map[Attribute]int
	Rows       int
	Cols       int
	ClassIndex int
	rowCount   int
	// How many rows do we allocate in one go?
	fixedAllocationSize      uint32
	fixedAllocationRowAmount uint32
	// Which row corresponds to which EdfRange?
	byteLevelMapping []edf.EdfRange
	// Where should we insert new entries?
	lastAllocation *edf.EdfRange
	// When was the last time we allocated?
	lastAllocatedAt int
	// How many rows in total do we have space for
	rowsAllocated int
}

func xorFloatOp(item float64) float64 {
	var ret float64
	var tmp int64
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, item)
	binary.Read(buf, binary.LittleEndian, &tmp)
	tmp ^= -1 << 63
	binary.Write(buf, binary.LittleEndian, tmp)
	binary.Read(buf, binary.LittleEndian, &ret)
	return ret
}

func printFloatByteArr(arr [][]byte) {
	buf := bytes.NewBuffer(nil)
	var f float64
	for _, b := range arr {
		buf.Write(b)
		binary.Read(buf, binary.LittleEndian, &f)
		f = xorFloatOp(f)
		fmt.Println(f)
	}
}

func (inst *Instances) resolveToInternal(attrs []Attribute) ([]int, error) {
	ret := make([]int, len(attrs))
	counter := 0
	selfAttrs := inst.GetAttrs()
	for i := range selfAttrs {
		for j := range attrs {
			if selfAttrs[i].Equals(attrs[j]) {
				ret[j] = i
				counter++
				break
			}
		}
	}
	if counter != len(attrs) {
		return nil, fmt.Errorf("Not all attributes resolved %s %s", attrs, selfAttrs)
	}
	return ret, nil
}

// Retrieves information about the Attributes attached to this
// set of Instances.
func (inst *Instances) GetAttrs() map[int]Attribute {
	ret := make(map[int]Attribute)
	for i, j := range inst.attributes {
		ret[i] = j
	}
	return ret
}

func (inst *Instances) GetAttr(i int) Attribute {
	return inst.attributes[i]
}

// GetClassAttrs returns the set of Attributes that are currently
// designated as class variables
func (inst *Instances) GetClassAttrs() map[int]Attribute {
	ret := make(map[int]Attribute)
	ret[inst.ClassIndex] = inst.attributes[inst.ClassIndex]
	return ret
}

// Sort does an in-place radix sort of Instances, using SortDirection
// direction (Ascending or Descending) with attrs as a slice of Attributes
// that you want to sort by.
//
// IMPORTANT: Radix sort is not stable, so ordering outside
// the attributes used for sorting is arbitrary.
func (inst *Instances) Sort(direction SortDirection, attributes []Attribute) error {
	attrs, err := inst.resolveToInternal(attributes)
	if err != nil {
		return err
	}
	// Create a buffer
	buf := bytes.NewBuffer(nil)
	ds := make([][]byte, inst.Rows)
	rs := make([]int, inst.Rows)
	for i := 0; i < inst.Rows; i++ {
		byteBuf := make([]byte, 8*len(attrs))
		for _, a := range attrs {
			x := inst.get(i, a)
			binary.Write(buf, binary.LittleEndian, xorFloatOp(x))
		}
		buf.Read(byteBuf)
		ds[i] = byteBuf
		rs[i] = i
	}
	// Sort values
	valueBins := make([][][]byte, 256)
	rowBins := make([][]int, 256)
	for i := 0; i < 8*len(attrs); i++ {
		for j := 0; j < len(ds); j++ {
			// Address each row value by it's ith byte
			b := ds[j]
			valueBins[b[i]] = append(valueBins[b[i]], b)
			rowBins[b[i]] = append(rowBins[b[i]], rs[j])
		}
		j := 0
		for k := 0; k < 256; k++ {
			bs := valueBins[k]
			rc := rowBins[k]
			copy(ds[j:], bs)
			copy(rs[j:], rc)
			j += len(bs)
			valueBins[k] = bs[:0]
			rowBins[k] = rc[:0]
		}
	}

	for _, b := range ds {
		var v float64
		buf.Write(b)
		binary.Read(buf, binary.LittleEndian, &v)
	}

	done := make([]bool, inst.Rows)
	for index := range rs {
		if done[index] {
			continue
		}
		j := index
		for {
			done[j] = true
			if rs[j] != index {
				inst.swapRows(j, rs[j])
				j = rs[j]
			} else {
				break
			}
		}
	}

	if direction == Descending {
		// Reverse the matrix
		for i, j := 0, inst.Rows-1; i < j; i, j = i+1, j-1 {
			inst.swapRows(i, j)
		}
	}
	return nil
}

// NewInstances returns a preallocated Instances structure
// with some helful values pre-filled.
func NewInstances(attrs []Attribute, rows int) *Instances {

	ret := new(Instances)

	ret.Rows = rows
	ret.Cols = len(attrs)
	ret.ClassIndex = len(attrs) - 1
	ret.rowCount = 0

	// Create Attribute mapping
	attrLookup := make(map[Attribute]int)
	for i, a := range attrs {
		attrLookup[a] = i
	}

	ret.attributes = attrs
	ret.attrLookup = attrLookup

	// Allocate storage
	storage, err := edf.EdfAnonMap()
	if err != nil {
		panic(err)
	}
	ret.storage = storage

	// Find out the optimal allocation size
	// Determine the length of each row (easy as we're still fixed
	// on 8-byte lengths)
	pageSize := uint32(storage.GetPageSize())
	rowLength := uint32(ret.computeRowStorageRequirements())
	// Then need to allocate the correct number of pages to minimize
	// wastage
	wastage := rowLength % pageSize
	for n := uint32(1); n < 4096 && ((n%pageSize)*wastage)%pageSize != 0; n++ {
		// Computing the optimum size
		ret.fixedAllocationSize = n
		ret.fixedAllocationRowAmount = (n * pageSize) / rowLength
	}
	ret.byteLevelMapping = make([]edf.EdfRange, 0)
	return ret
}

func (inst *Instances) computeRowStorageRequirements() int {
	return 8 * len(inst.attributes)
}

func (inst *Instances) extendAllocation() error {
	// Allocate more space
	r, err := inst.storage.AllocPages(inst.fixedAllocationSize, 2) // Allocate to the FIXED thread
	if err != nil {
		return fmt.Errorf("No space available or (%s)", err)
	}
	// Get the total size of the allocation
	size := r.Size()
	// Compute the number of additional rows
	rowLength := inst.computeRowStorageRequirements()
	additionalRows := int(size) / rowLength
	// Cache the EdfRange for future reference
	inst.byteLevelMapping = append(inst.byteLevelMapping, r)
	inst.lastAllocatedAt = inst.rowCount
	// Add that to the rows available
	inst.lastAllocation = &r
	inst.rowsAllocated += additionalRows
	return nil
}

// GetAttrIndex returns the index of the first matching Attribute
func (inst *Instances) GetAttrIndex(a Attribute) int {
	for attrIndex, attr := range inst.attributes {
		if attr.Equals(a) {
			return attrIndex
		}
	}
	return -1
}

// CountAttrValues returns the distribution of values of a given
// Attribute.
// IMPORTANT: calls panic() if the attribute index of a cannot be
// determined. Call GetAttrIndex(a) and check for a -1 return value.
// STATUS: Compatable
func (inst *Instances) CountAttrValues(a Attribute) map[string]int {
	ret := make(map[string]int)

	// Find the attribute index for this value
	attrIndex := inst.GetAttrIndex(a)

	if attrIndex == -1 {
		panic("Invalid attribute")
	}
	for i := 0; i < inst.Rows; i++ {
		sysVal := inst.get(i, attrIndex)
		convVal := PackFloatToBytes(sysVal)
		stringVal := a.GetStringFromSysVal(convVal)
		ret[stringVal] += 1
	}
	return ret
}

// DecomposeOnAttributeValues divides the instance set depending on the
// value of a given Attribute, constructs child instances, and returns
// them in a map keyed on the string value of that Attribute.
// IMPORTANT: calls panic() if the attribute index of at cannot be determined.
// Use GetAttrIndex(at) and check for a non-zero return value.
// STATUS: Compatable
func (inst *Instances) DecomposeOnAttributeValues(at Attribute) map[string]UpdatableDataGrid {
	// Find the attribute we're decomposing on
	attrIndex := inst.GetAttrIndex(at)
	if attrIndex == -1 {
		panic("Invalid attribute index")
	}
	// Construct the new attribute set
	newAttrs := make([]Attribute, 0)
	for i := range inst.attributes {
		a := inst.attributes[i]
		if a.Equals(at) {
			continue
		}
		newAttrs = append(newAttrs, a)
	}
	// Create the return map, several counting maps
	ret := make(map[string]UpdatableDataGrid)
	counts := inst.CountAttrValues(at) // So we know what to allocate
	for k := range counts {
		tmp := NewInstances(newAttrs, counts[k])
		ret[k] = tmp
	}

	// Range over and filter the attributes, add to appropriate row
	inst.MapOverRowsExplicit(inst.GetAttrs(), func(row map[Attribute][]byte, i int) (bool, error) {

		attrVal := at.GetStringFromSysVal(row[at])
		delete(row, at)
		err := ret[attrVal].AppendRowExplicit(row)
		if err != nil {
			panic(err)
		}
		return true, nil
	})

	return ret
}

// GetRowExplicit returns a map containing the values of the selected Attributes
// at a particular row.
func (inst *Instances) GetRowExplicit(attrs map[int]Attribute, row int) map[Attribute][]byte {
	ret := make(map[Attribute][]byte)
	// Eliminate the call to get in this function
	for a := range attrs {
		ret[attrs[a]] = PackFloatToBytes(inst.get(row, a))
	}
	return ret
}

func intraMappingCopy(mappings [][]byte, dest []byte, startOffset int) {
	copyLen := len(dest)
	byteCounter := 0
	if startOffset < len(mappings[0]) {
		// Case 1: copy from the source offset mappings[0]
		// until the end of the source offset
		copySrc := mappings[0][startOffset:]
		byteCounter += copy(dest, copySrc)
	}
	if len(mappings[1]) > 0 {
		// Case 2: copy any stuff at the start of mappings[1]
		// until the end of copyLen
		copy(dest[byteCounter:], mappings[1][:copyLen-byteCounter])
	}
}

func (inst *Instances) AllocRowVector(attrs map[int]Attribute) ([][]byte, []int) {
	ret := make([][]byte, len(attrs))
	retAttr := make([]int, len(attrs))
	for i := 0; i < len(attrs); i++ {
		ret[i] = make([]byte, 8)
	}
	counter := 0
	for i := range attrs {
		retAttr[counter] = i
		counter++
	}
	return ret, retAttr
}

func (inst *Instances) GetRow(row int, attrs []int, out [][]byte) int {
	// With 8-byte lengths, maximum span of mappings is two
	// Need an AllocRowVector
	var s1 []byte
	var s2 []byte

	// Translate the row into an allocation
	rowAlloc := row / int(inst.fixedAllocationRowAmount)
	rowOffset := row % int(inst.fixedAllocationRowAmount)
	rowLength := 8 * len(inst.attributes)
	rowRange := inst.byteLevelMapping[rowAlloc]

	// Resolve the primary block
	s1, next := inst.storage.IResolveRange(rowRange, 0)
	// Resolve the next block (if relevant)
	if next != 0 {
		s2, _ = inst.storage.IResolveRange(rowRange, next)
	} else {
		s2 = nil
	}

	colCounter := 0
	for col := range attrs {
		s1ColOffset := rowOffset * rowLength + col * 8
		if next != 0 {
			byteCounter := 0
			if s1ColOffset < len(s1) {
				src := s1[s1ColOffset:]
				if len(src) > 0 {
					byteCounter += copy(out[colCounter], src)
				}
			}
			copy(out[colCounter][byteCounter:], s2[:8-byteCounter])
		} else {
			copy(out[colCounter], s1[s1ColOffset:])
		}
		colCounter++
	}
	return 0
}

// AddAttribute adds a new Attribute to this set of Instances
// IMPORTANT: will return an error code if this set of instances
// has been allocated (and leave the Attribute set unmodified)
func (inst *Instances) AddAttribute(a Attribute) error {
	if inst.storage != nil {
		return fmt.Errorf("Can't resize online")
	}
	inst.attributes = append(inst.attributes, a)
	inst.attrLookup[a] = len(inst.attributes) - 1
	return nil
}

// RemoveAttribute removes an Attribute from this set of Instances
// IMPORTANT: will return an error code if this set of instances
// has been allocated (and will also leave the Attribute set unmodified)
func (inst *Instances) RemoveAttribute(a Attribute) error {
	if inst.storage != nil {
		return fmt.Errorf("Can't resize online!")
	}
	revisedAttributes := make([]Attribute, 0)
	for i := range inst.attributes {
		if inst.attributes[i].Equals(a) {
			continue
		}
		revisedAttributes = append(revisedAttributes, inst.attributes[i])
	}
	inst.attributes = revisedAttributes
	return nil
}

// AppendRowExplicit adds the given row map to this set of Instances.
// Allocates the required storage if unallocated.
// IMPORTANT: will return an error code (and won't add the row)
// if a) the number of rows exceeds the space allocated
// b) if the row map contains values with more than 8 bytes
// c) if the row map contains unrecognised Attributes
// TODO: Update positionMap generation
func (inst *Instances) AppendRowExplicit(row map[Attribute][]byte) error {
	// Convert attributes into offsets
	positionMap := make([][]byte, len(row))
	for a := range row {
		pos, ok := inst.attrLookup[a]
		if !ok {
			return fmt.Errorf("Couldn't resolve attribute %s", a)
		}
		positionMap[pos] = row[a]
	}
	return inst.AppendRow(positionMap)
}

// AppendRow adds the given row map to this set of Instances.
// Allocates the required storage if unallocated.
// IMPORTANT: will return an error code (and won't add the row)
// if a) the number of rows exceeds the space allocated
// b) if the row map contains values with more than 8 bytes
// c) if the row map contains unrecognised Attributes
func (inst *Instances) AppendRow(row [][]byte) error {
	// Double check that we've got enough space allocated
	if inst.rowsAllocated <= inst.rowCount {
		err := inst.extendAllocation()
		if err != nil {
			return fmt.Errorf("Row allocation failure: %s", err)
		}
	}
	//
	// Convert bytes into values, store in matrix
	for col := range row {
		valf := UnpackBytesToFloat(row[col])
		inst.set(inst.rowCount, col, valf)
	}
	inst.rowCount++
	return nil
}

// MapOverRowsExplicit passes each row map into a function used for training
// Within the closure, return `false, nil` to indicate the end of
// processing, or return `_, error` to indicate a problem.
func (inst *Instances) MapOverRowsExplicit(attrs map[int]Attribute, mapFunc func(map[Attribute][]byte, int) (bool, error)) error {
	for i := 0; i < inst.Rows; i++ {
		row := inst.GetRowExplicit(attrs, i)
		ok, err := mapFunc(row, i)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

func (inst *Instances) MapOverRows(attrs map[int]Attribute, mapFunc func([][]byte, int) (bool, error)) error {
	rowBuf, attrArr := inst.AllocRowVector(attrs)
	for i := 0; i < inst.Rows; i++ {
		inst.GetRow(i, attrArr, rowBuf)
		ok, err := mapFunc(rowBuf, i)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

// GetClassDistribution returns the class distribution after a hypothetical split
// STATUS: Compatable
func (inst *Instances) GetClassDistributionAfterSplit(at Attribute) map[string]map[string]int {

	ret := make(map[string]map[string]int)

	// Find the attribute we're decomposing on
	attrIndex := inst.GetAttrIndex(at)
	if attrIndex == -1 {
		panic("Invalid attribute index")
	}

	// Get the class index
	classAttr := inst.attributes[inst.ClassIndex]

	for i := 0; i < inst.Rows; i++ {
		sysVal := inst.get(i, attrIndex)
		convVal := PackFloatToBytes(sysVal)
		splitVar := at.GetStringFromSysVal(convVal)
		sysVal = inst.get(i, inst.ClassIndex)
		convVal = PackFloatToBytes(sysVal)
		classVar := classAttr.GetStringFromSysVal(convVal)
		if _, ok := ret[splitVar]; !ok {
			ret[splitVar] = make(map[string]int)
			i--
			continue
		}
		ret[splitVar][classVar]++
	}

	return ret
}

//
// Internal access functions
//

// get returns the system representation (float64) of the value
// stored at the given row and col coordinate.
// get is the single biggest performance problem
// have to make it faster!
func (inst *Instances) get(row int, col int) float64 {
	// Translate the row into an allocation
	rowAlloc := row / int(inst.fixedAllocationRowAmount)
	rowOffset := row % int(inst.fixedAllocationRowAmount)
	rowLength := 8 * len(inst.attributes)
	// Translate the column into a position
	col *= 8
	// Get the byte offsets
	rawData := inst.storage.ResolveRange(inst.byteLevelMapping[rowAlloc])
	if len(rawData) == 1 {
		return UnpackBytesToFloat(rawData[0][rowOffset*rowLength+col:])
	} else if len(rawData) == 2 {
		return 0 // TODO: Write code to handle this case
	} else {
		panic("segment")
	}
	return 0
}

// set sets the system representation (float64) to val at the
// given row and column coordinate.
func (inst *Instances) set(row int, col int, val float64) {
	// Translate the row into an allocation
	rowAlloc := row / int(inst.fixedAllocationRowAmount)
	rowOffset := row % int(inst.fixedAllocationRowAmount)
	rowLength := 8 * len(inst.attributes)
	// Translate the column into a position
	col *= 8
	// Get the byte offsets
	//	fmt.Printf("%d %d %d\n", row, int(inst.fixedAllocationRowAmount), rowAlloc)
	mappingRange := inst.byteLevelMapping[rowAlloc]
	rawData := inst.storage.ResolveRange(mappingRange)
	if len(rawData) == 1 {
		PackFloatToBytesInline(val, rawData[0][rowOffset*rowLength+col:])
	} else if len(rawData) == 2 {
		// TODO: write code to handle this case
	} else {
		panic("segment")
	}
}

//
// Printing functions
//

// RowStr returns a human-readable representation of a given row.
// STATUS: Compatable
func (inst *Instances) RowStr(row int) string {
	// Prints a given row
	var buffer bytes.Buffer
	for j := 0; j < inst.Cols; j++ {
		val := inst.get(row, j)
		convVal := PackFloatToBytes(val)
		a := inst.attributes[j]
		postfix := " "
		if j == inst.Cols-1 {
			postfix = ""
		}

		buffer.WriteString(fmt.Sprintf("%s%s", a.GetStringFromSysVal(convVal), postfix))
	}
	return buffer.String()
}

// String returns a human-readable representation of this dataset
// STATUS: Compatable
func (inst *Instances) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("Instances with ")
	buffer.WriteString(fmt.Sprintf("%d row(s) ", inst.Rows))
	buffer.WriteString(fmt.Sprintf("%d attribute(s)\n", inst.Cols))

	buffer.WriteString(fmt.Sprintf("Attributes: \n"))
	for i, a := range inst.attributes {
		prefix := "\t"
		if i == inst.ClassIndex {
			prefix = "*\t"
		}
		buffer.WriteString(fmt.Sprintf("%s%s\n", prefix, a))
	}

	buffer.WriteString("\nData:\n")
	maxRows := 30
	if inst.Rows < maxRows {
		maxRows = inst.Rows
	}

	for i := 0; i < maxRows; i++ {
		buffer.WriteString("\t")
		for j := 0; j < inst.Cols; j++ {
			val := inst.get(i, j)
			convVal := PackFloatToBytes(val)
			a := inst.attributes[j]
			buffer.WriteString(fmt.Sprintf("%s ", a.GetStringFromSysVal(convVal)))
		}
		buffer.WriteString("\n")
	}

	missingRows := inst.Rows - maxRows
	if missingRows != 0 {
		buffer.WriteString(fmt.Sprintf("\t...\n%d row(s) undisplayed", missingRows))
	} else {
		buffer.WriteString("All rows displayed")
	}

	return buffer.String()
}

// SelectAttributes returns a new instance set containing
// the values from this one with only the Attributes specified
func (inst *Instances) SelectAttributes(attrs []Attribute) DataGrid {
	ret := NewInstances(attrs, inst.Rows)
	attrIndices := make([]int, 0)
	for _, a := range attrs {
		attrIndex := inst.GetAttrIndex(a)
		attrIndices = append(attrIndices, attrIndex)
	}
	for i := 0; i < inst.Rows; i++ {
		for j, a := range attrIndices {
			ret.set(i, j, inst.get(i, a))
		}
	}
	return ret
}

// Size returns the current size of this set of Instances
// order is column, then row
func (inst *Instances) Size() (uint64, uint64) {
	return uint64(len(inst.attributes)), uint64(inst.rowCount)
}

// Shuffle randomizes the row order in place
func (inst *Instances) Shuffle() FixedDataGrid {
	for i := 0; i < inst.Rows; i++ {
		j := rand.Intn(i + 1)
		inst.swapRows(i, j)
	}
	return inst
}

// Equal checks whether a given Instance set is exactly the same
// as another: same size and same values (as determined by the Attributes)
//
// IMPORTANT: does not explicitly check if the Attributes are considered equal.
// Status: Compatable
func (inst *Instances) Equals(otherGrid DataGrid) bool {
	other, ok := otherGrid.(*Instances)
	if !ok {
		return false
	}
	if inst.Rows != other.Rows {
		return false
	}
	if inst.Cols != other.Cols {
		return false
	}
	for i := 0; i < inst.Rows; i++ {
		eq := inst.RowStr(i) == other.RowStr(i)
		if !eq {
			return false
		}
	}
	return true
}

func (inst *Instances) swapRows(r1 int, r2 int) {

	rowLength := 8 * len(inst.attributes)

	// Translate r1 into an allocation
	rowAlloc := r1 / int(inst.fixedAllocationRowAmount)
	rowOffset := r1 % int(inst.fixedAllocationRowAmount)
	// Get the byte offsets for r1
	r1Data := inst.storage.ResolveRange(inst.byteLevelMapping[rowAlloc])
	r1Slice := r1Data[0][rowOffset*rowLength : (rowOffset+1)*rowLength]

	// Translate r2 into an allocation
	rowAlloc = r2 / int(inst.fixedAllocationRowAmount)
	rowOffset = r2 % int(inst.fixedAllocationRowAmount)
	// Get the byte offsets for r2
	r2Data := inst.storage.ResolveRange(inst.byteLevelMapping[rowAlloc])
	r2Slice := r2Data[0][rowOffset*rowLength : (rowOffset+1)*rowLength]

	// Create a temporary buffer for copying
	buf := make([]byte, rowLength)

	copy(buf, r2Slice)
	copy(r2Slice, r1Slice)
	copy(r1Slice, buf)
}
