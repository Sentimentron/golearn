package filters

import (
	"github.com/sjwhitworth/golearn/base"
)

// CategoricalToBinaryFilter replaces a CategoricalAttribute
// with a Binary
type ClassBinaryFilter struct {
	attrs        map[base.Attribute]base.Attribute
	classAttr    base.Attribute
	classAttrVal uint64
}

func (f *oneVsAllFilter) AddAttribute(a base.Attribute) error {
	return fmt.Errorf("Not supported")
}

func (f *oneVsAllFilter) GetAttributesAfterFiltering() []base.FilteredAttribute {
	ret := make([]base.FilteredAttribute, len(f.attrs))
	cnt := 0
	for i := range f.attrs {
		ret[cnt] = base.FilteredAttribute{i, f.attrs[i]}
		cnt++
	}
	return ret
}
func (f *oneVsAllFilter) String() string {
	return "oneVsAllFilter"
}
func (f *oneVsAllFilter) Transform(old, to base.Attribute, seq []byte) []byte {
	if !old.Equals(f.classAttr) {
		return seq
	}
	val := base.UnpackBytesToU64(seq)
	if val == f.classAttrVal {
		return base.PackFloatToBytes(1.0)
	}
	return base.PackFloatToBytes(0.0)
}
func (f *oneVsAllFilter) Train() error {
	return fmt.Errorf("Unsupported")
}
