package base

// SortDirection specifies sorting direction...
type SortDirection int

const (
	// Descending says that Instances should be sorted high to low...
	Descending SortDirection = 1
	// Ascending states that Instances should be sorted low to high...
	Ascending SortDirection = 2
)

type DataGrid interface {
	// Pass all rows into a training closure
	MapOverRowsExplicit(map[int]Attribute, func(map[Attribute][]byte, int) (bool, error)) error
	MapOverRows(map[int]Attribute, func([][]byte, int) (bool, error)) error
	// Returns a single Attribute with a given index
	GetAttr(int) Attribute
	// Returns a int->Attribute map. Shouldn't be incompatibly changed.
	GetAttrs() map[int]Attribute
	// Returns a int->Attribute map containing classes
	GetClassAttrs() map[int]Attribute
	// Returns a new set of instances containing only the selected columns
	SelectAttributes(attrs []Attribute) DataGrid
	// Returns a human readable string
	String() string
	// Checks if two DataGrids are equal
	Equals(DataGrid) bool
}

type FixedDataGrid interface {
	DataGrid
	// Divides the instance set depending on the value of a
	// given Attribute
	DecomposeOnAttributeValues(Attribute) map[string]UpdatableDataGrid
	// Sorts the DataGrid in place on the given attribute
	// Not supported by all DataGrid implementations
	Sort(SortDirection, []Attribute) error
	// Returns the distribution of values of a given Attribute
	CountAttrValues(Attribute) map[string]int
	// GetClassDistributionAfterSplit returns the class distributuion
	// after a hypothetical split
	GetClassDistributionAfterSplit(Attribute) map[string]map[string]int
	// RowStr returns the string representation of a given row
	RowStr(int) string
	// GetRow returns the GetRows() response at a given row
	// GetRow(map[int]Attribute, int) [][]byte
        GetRow(int, map[int]Attribute, [][]byte) int
	// GetRowExplicit
	GetRowExplicit(map[int]Attribute, int) map[Attribute][]byte
	// Shuffle randomizes the row order
	Shuffle() FixedDataGrid
	// Size returns two uint64s, column count and then row
	Size() (uint64, uint64)
}

type UpdatableDataGrid interface {
	FixedDataGrid
	// AppendRow inserts a new row
	AppendRow([][]byte) error
	// AppendRowExplicit inserts a new row
	AppendRowExplicit(map[Attribute][]byte) error
	// Add a new attribute
	AddAttribute(Attribute) error
	// Delete an attribute
	RemoveAttribute(Attribute) error
}
