package base

// FilteredAttributes represent a mapping from the output
// generated by a filter to the original value.
type FilteredAttribute struct {
	Old Attribute
	New Attribute
}

// Filters transform the byte sequences stored in DataGrid
// implementations.
type Filter interface {
	// Adds an Attribute to the filter
	AddAttribute(Attribute) error
	// Allows mapping old to new Attributes
	GetAttributesAfterFiltering() []FilteredAttribute
	// Accepts an old Attribute and a byte sequence
	Transform(Attribute, []byte) []byte
	// Builds the filter
	Train() error
}