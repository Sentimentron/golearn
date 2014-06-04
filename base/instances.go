package base

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gonum/matrix/mat64"
	"math/rand"
	"reflect"
)

// Instances represents a grid of numbers (typed by Attributes)
// stored internally in mat.DenseMatrix as float64's.
// See docs/instances.md for more information.
type Instances struct {
	storage    *mat64.Dense
	attributes []Attribute
	Rows       int
	Cols       int
	ClassIndex int
	rowCount   int
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
	fmt.Println(attrs)
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
			x := inst.storage.At(i, a)
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
	rawStorage := make([]float64, rows*len(attrs))
	return NewInstancesFromRaw(attrs, rows, rawStorage)
}

// CheckNewInstancesFromRaw checks whether a call to NewInstancesFromRaw
// is likely to produce an error-free result.
func CheckNewInstancesFromRaw(attrs []Attribute, rows int, data []float64) error {
	size := rows * len(attrs)
	if size < len(data) {
		return errors.New("base: data length is larger than the rows * attribute space.")
	} else if size > len(data) {
		return errors.New("base: data is smaller than the rows * attribute space")
	}
	return nil
}

// NewInstancesFromRaw wraps a slice of float64 numbers in a
// mat64.Dense structure, reshaping it with the given number of rows
// and representing it with the given attrs (Attribute slice)
//
// IMPORTANT: if the |attrs| * |rows| value doesn't equal len(data)
// then panic()s may occur. Use CheckNewInstancesFromRaw to confirm.
func NewInstancesFromRaw(attrs []Attribute, rows int, data []float64) *Instances {
	rawStorage := mat64.NewDense(rows, len(attrs), data)
	return NewInstancesFromDense(attrs, rows, rawStorage)
}

// NewInstancesFromDense creates a set of Instances from a mat64.Dense
// matrix
func NewInstancesFromDense(attrs []Attribute, rows int, mat *mat64.Dense) *Instances {
	return &Instances{mat, attrs, rows, len(attrs), len(attrs) - 1, 0}
}

// InstancesTrainTestSplit takes a given Instances (src) and a train-test fraction
// (prop) and returns an array of two new Instances, one containing approximately
// that fraction and the other containing what's left.
//
// IMPORTANT: this function is only meaningful when prop is between 0.0 and 1.0.
// Using any other values may result in odd behaviour.
func InstancesTrainTestSplit(src *Instances, prop float64) (*Instances, *Instances) {
	trainingRows := make([]int, 0)
	testingRows := make([]int, 0)
	numAttrs := len(src.attributes)
	src.Shuffle()
	for i := 0; i < src.Rows; i++ {
		trainOrTest := rand.Intn(101)
		if trainOrTest > int(100*prop) {
			trainingRows = append(trainingRows, i)
		} else {
			testingRows = append(testingRows, i)
		}
	}

	rawTrainMatrix := mat64.NewDense(len(trainingRows), numAttrs, make([]float64, len(trainingRows)*numAttrs))
	rawTestMatrix := mat64.NewDense(len(testingRows), numAttrs, make([]float64, len(testingRows)*numAttrs))

	for i, row := range trainingRows {
		rowDat := src.storage.RowView(row)
		rawTrainMatrix.SetRow(i, rowDat)
	}
	for i, row := range testingRows {
		rowDat := src.storage.RowView(row)
		rawTestMatrix.SetRow(i, rowDat)
	}

	trainingRet := NewInstancesFromDense(src.attributes, len(trainingRows), rawTrainMatrix)
	testRet := NewInstancesFromDense(src.attributes, len(testingRows), rawTestMatrix)
	return trainingRet, testRet
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
func (inst *Instances) DecomposeOnAttributeValues(at Attribute) map[string]*Instances {
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
	ret := make(map[string]*Instances)
	counts := inst.CountAttrValues(at) // So we know what to allocate
	rows := make(map[string]int)
	for k := range counts {
		tmp := NewInstances(newAttrs, counts[k])
		ret[k] = tmp
	}
	for i := 0; i < inst.Rows; i++ {
		newAttrCounter := 0
		sysVal := inst.get(i, attrIndex)
		convVal := PackFloatToBytes(sysVal)
		classVar := at.GetStringFromSysVal(convVal)
		dest := ret[classVar]
		destRow := rows[classVar]
		for j := 0; j < inst.Cols; j++ {
			a := inst.attributes[j]
			if a.Equals(at) {
				continue
			}
			dest.set(destRow, newAttrCounter, inst.get(i, j))
			newAttrCounter++
		}
		rows[classVar]++
	}
	return ret
}

// GetRow returns a map containing the values of the selected Attributes
// at a particular row.
func (inst *Instances) GetRow(attrs map[int]Attribute, row int) map[Attribute][]byte {
	ret := make(map[Attribute][]byte)
	for a := range attrs {
		ret[attrs[a]] = PackFloatToBytes(inst.get(row, a))
	}
	return ret
}

// AddAttribute adds a new Attribute to this set of Instances
// IMPORTANT: will return an error code if this set of instances
// has been allocated (and leave the Attribute set unmodified)
func (inst *Instances) AddAttribute(a Attribute) error {
	if inst.storage != nil {
		return fmt.Errorf("Can't resize online")
	}
	inst.attributes = append(inst.attributes, a)
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

// AppendRow adds the given row map to this set of Instances.
// Allocates the required storage if unallocated.
// IMPORTANT: will return an error code (and won't add the row)
// if a) the number of rows exceeds the space allocated
// b) if the row map contains values with more than 8 bytes
// c) if the row map contains unrecognised Attributes
func (inst *Instances) AppendRow(row map[Attribute][]byte) error {
	// If we haven't allocated yet...
	if inst.storage == nil {
		// Allocate new storage
		tmp := make([]float64, inst.Rows*len(inst.attributes))
		inst.storage = mat64.NewDense(inst.Rows, len(inst.attributes), tmp)
	}
	// Double check that we've got enough space allocated
	if inst.Rows <= inst.rowCount {
		return fmt.Errorf("No space available")
	}
	// Convert attributes into offsets
	positionMap := make(map[int][]byte)
	for a := range row {
		matched := -1
		for i := range inst.attributes {
			if inst.attributes[i].Equals(a) {
				matched = i
				break
			}
		}
		if matched == -1 {
			return fmt.Errorf("Couldn't resolve attribute %s", a)
		}
		if len(row[a]) != 8 {
			return fmt.Errorf("Variable width types aren't supported")
		}
		positionMap[matched] = row[a]
	}
	// Convert bytes into values, store in matrix
	for col := range positionMap {
		valf := UnpackBytesToFloat(positionMap[col])
		inst.set(inst.rowCount, col, valf)
	}
	inst.rowCount++
	return nil
}

// MapOverRows passes each row map into a function used for training
// Within the closure, return `false, nil` to indicate the end of
// processing, or return `_, error` to indicate a problem.
func (inst *Instances) MapOverRows(attrs map[int]Attribute, mapFunc func(map[Attribute][]byte) (bool, error)) error {
	for i := 0; i < inst.Rows; i++ {
		row := inst.GetRow(attrs, i)
		ok, err := mapFunc(row)
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
func (inst *Instances) get(row int, col int) float64 {
	return inst.storage.At(row, col)
}

// set sets the system representation (float64) to val at the
// given row and column coordinate.
func (inst *Instances) set(row int, col int, val float64) {
	//fmt.Printf("%d %d %.2f\n", row, col, val)
	inst.storage.Set(row, col, val)
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
		val := inst.storage.At(row, j)
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
			val := inst.storage.At(i, j)
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

// Shuffle randomizes the row order in place
func (inst *Instances) Shuffle() {
	for i := 0; i < inst.Rows; i++ {
		j := rand.Intn(i + 1)
		inst.swapRows(i, j)
	}
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
		row1 := inst.GetRow(inst.GetAttrs(), i)
		row2 := other.GetRow(inst.GetAttrs(), i)
		eq := reflect.DeepEqual(row1, row2)
		if !eq {
			return false
		}
	}
	return true
}

func (inst *Instances) swapRows(r1 int, r2 int) {
	row1buf := make([]float64, inst.Cols)
	row2buf := make([]float64, inst.Cols)
	row1 := inst.storage.RowView(r1)
	row2 := inst.storage.RowView(r2)
	copy(row1buf, row1)
	copy(row2buf, row2)
	inst.storage.SetRow(r1, row2buf)
	inst.storage.SetRow(r2, row1buf)
}
