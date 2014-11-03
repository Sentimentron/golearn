package base

import (
	"fmt"
)

// StringAttribute is an Attribute storing raw text.
type StringAttribute struct {
	Name string
	vals []string
}

// NewStringAttribute creates a StringAttribute with a given name.
func NewStringAttribute(name string) *StringAttribute {
	return &StringAttribute{name, make([]string, 0)}
}

// GetType returns something random.
func (a *StringAttribute) GetType() int {
	return CategoricalType
}

// GetName return the current name for this IDAttribute.
func (a *StringAttribute) GetName() string {
	return a.Name
}

// SetName sets the current name of this IDAttribute.
func (a *StringAttribute) SetName(name string) {
	a.Name = name
}

// String gets a human-readable version of this IDAttribute.
func (a *StringAttribute) String() string {
	return fmt.Sprintf("StringAttribute(%s, %v)", a.Name, len(a.vals))
}

// GetSysValFromString returns the byte sequence which represents
// the given string.
func (a *StringAttribute) GetSysValFromString(str string) []byte {
	a.vals = append(a.vals, str)
	return PackU64ToBytes(uint64(len(a.vals) - 1))
}

// GetStringFromSysVal returns a string from a given value.
func (a *StringAttribute) GetStringFromSysVal(val []byte) string {
	v := UnpackBytesToU64(val)
	return a.vals[v]
}

// Equals checks if this and another Attribute are StringAttributes
// and have the same name.
func (a *StringAttribute) Equals(other Attribute) bool {
	if attr, ok := other.(*StringAttribute); ok {
		return a.Name == attr.Name
	}
	return false
}

// Compatible checks whether an Attribute is either another
// IDAttribute or a CategoricalAttribute.
func (a *StringAttribute) Compatible(other Attribute) bool {
	if _, ok := other.(*IDAttribute); ok {
		return true
	}
	if _, ok := other.(*StringAttribute); ok {
		return true
	}
	if _, ok := other.(*CategoricalAttribute); ok {
		return true
	}

	return false
}
