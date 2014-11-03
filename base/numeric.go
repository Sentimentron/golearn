package base

import (
	"fmt"
	"strconv"
)

// IDAttribute is designed to link rows in a datagrid
// with those coming from a database. It's supposed to be ignored
// by all classification algorithms and pass unmodified.
type IDAttribute struct {
	Name string
}

// GetType returns something random.
func (a *IDAttribute) GetType() int {
	return CategoricalType
}

// GetName return the current name for this IDAttribute.
func (a *IDAttribute) GetName() string {
	return a.Name
}

// SetName sets the current name of this IDAttribute.
func (a *IDAttribute) SetName(name string) {
	a.Name = name
}

// String gets a human-readable version of this IDAttribute.
func (a *IDAttribute) String() string {
	return fmt.Sprintf("IDAttribute(%s)", a.Name)
}

// GetSysValFromString returns the byte sequence which represents
// the given string.
func (a *IDAttribute) GetSysValFromString(str string) []byte {
	val, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return nil
	}
	return PackU64ToBytes(val)
}

// GetStringFromSysVal returns a string from a given value.
func (a *IDAttribute) GetStringFromSysVal(val []byte) string {
	v := UnpackBytesToU64(val)
	return fmt.Sprintf("%d", v)
}

// Equals checks if this and another Attribute are IDAttributes
// and have the same name.
func (a *IDAttribute) Equals(other Attribute) bool {
	if attr, ok := other.(*IDAttribute); ok {
		return a.Name == attr.Name
	}
	return false
}

// Compatible checks whether an Attribute is either another
// IDAttribute or a CategoricalAttribute.
func (a *IDAttribute) Compatible(other Attribute) bool {
	if _, ok := other.(*IDAttribute); ok {
		return true
	}
	if _, ok := other.(*CategoricalAttribute); ok {
		return true
	}

	return false
}
