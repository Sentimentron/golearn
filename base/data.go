package base

// SortDirection specifies sorting direction...
type SortDirection int

const (
	// Descending says that Instances should be sorted high to low...
	Descending SortDirection = 1
	// Ascending states that Instances should be sorted low to high...
	Ascending SortDirection = 2
)

type FixedDataGrid interface {
	// Divides the instance set depending on the value of a
	// given Attribute
	DecomposeOnAttributeValues(Attribute) map[string]*Instances
	// Sorts the DataGrid in place on the given attribute
	// Not supported by all DataGrid implementations
	Sort(SortDirection, []Attribute) error
	// Returns the distribution of values of a given Attribute
	CountAttrValues(Attribute) map[string]int
	// GetClassDistributionAfterSplit returns the class distributuion
	// after a hypothetical split
	GetClassDistributionAfterSplit(Attribute) map[string]map[string]int
}

type DataGrid interface {
	// GetRows returns a channel containing int -> bytes maps
	// containing the attribute's rows for the selected Attributes
	GetRows([]Attribute) chan map[int][]byte
	// AppendRow inserts a new row
	AppendRow(map[int][]byte)
	// Returns a int->Attribute map. Shouldn't be incompatibly changed.
	GetAttrs() map[int]Attribute
	// Returns a int->Attribute map containing classes
	GetClassAttrs() map[int]Attribute
	// Add a new attribute
	AddAttribute(Attribute) error
	// Delete an attribute
	DeleteAttribute(Attribute) error
}
