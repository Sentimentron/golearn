package time

import (
	"fmt"
	"github.com/sjwhitworth/golearn/base"
	"time"
)

// EpochNSTimeAttribute represents time using the number
// of seconds since the epoch (Jan 1, 1970 UTC)
type EpochNSTimeAttribute struct {
	Name   string
	Layout string
}

// NewEpochNSTimeAttribute creates a new EpochNSTimeAttribute which
// represents dates with the given layout and has the given name.
func NewEpochNSTimeAttribute(name, layout string) *EpochNSTimeAttribute {
	return &EpochNSTimeAttribute{
		name,
		layout,
	}
}

// GetType returns 0 for EpochNSTimeAttribute
func (a *EpochNSTimeAttribute) GetType() int {
	// Stubbed
	return 0
}

// GetName returns this Attribute's current name
func (a *EpochNSTimeAttribute) GetName() string {
	return a.Name
}

// SetName sets this Attribute's current name
func (a *EpochNSTimeAttribute) SetName(n string) {
	a.Name = n
}

// GetSysValFromString parses the given date with the layout
// given for this Attribute.
//
// IMPORTANT: panics if the string cannot be parsed.
func (a *EpochNSTimeAttribute) GetSysValFromString(s string) []byte {
	t, err := time.Parse(a.Layout, s)
	if err != nil {
		panic(err)
	}
	t = t.UTC()
	return base.PackU64ToBytes(uint64(t.Unix()))
}

// GetStringFromSysVal produces the date string from a given system
// byte sequence using layout.
func (a *EpochNSTimeAttribute) GetStringFromSysVal(v []byte) string {
	u := base.UnpackBytesToU64(v)
	t := time.Unix(int64(u), 0.0).UTC()
	return t.Format(a.Layout)
}

// Equals checks that two EpochNSTimeAttributes are equivalent.
func (a *EpochNSTimeAttribute) Equals(other base.Attribute) bool {
	if o, ok := other.(*EpochNSTimeAttribute); ok {
		if o.Name == a.Name {
			return true
		}
	}
	return false
}

// Compatable checks whether this Attribute can be grouped with
// any others.
func (a *EpochNSTimeAttribute) Compatable(other base.Attribute) bool {
	if _, ok := other.(*base.BinaryAttribute); ok {
		return false
	}
	return true
}

// String returns a human-readable summary of this Attribute.
func (a *EpochNSTimeAttribute) String() string {
	return fmt.Sprintf("EpochNSTimeAttribute(%s, %s)", a.Name, a.Layout)
}
