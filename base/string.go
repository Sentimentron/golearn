package base

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

// StringAttribute holds variable length strings.
type StringAttribute struct {
	Name string
}

// NewStringAttribute creates a StringAttribute with a given name.
func NewStringAttribute(s string) *StringAttribute {
	ret := new(StringAttribute)
	ret.SetName(s)
	return ret
}

// MarshalJSON returns a JSON representation of this Attribute
// for serialisation.
func (s *StringAttribute) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type": "string",
		"name": s.Name,
	})
}

// UnmarhsalJSON reads the JSON representation of this Attribute.
func (s *StringAttribute) UnmarshalJSON(data []byte) error {
	var d map[string]interface{}
	err := json.Unmarshal(data, &d)
	if err != nil {
		return err
	}
	return nil
}

// GetName returns the name set for this StringAttribute.
func (s *StringAttribute) GetName() string {
	return s.Name
}

// SetName sets the name of this StringAttribute.
func (s *StringAttribute) SetName(n string) {
	s.Name = n
}

// String returns this Attribute in human-readable form.
func (s *StringAttribute) String() string {
	return fmt.Sprintf("StringAttribute(%s)", s.Name)
}

// GetSysVal returns the byte package of this thing.
func (s *StringAttribute) GetSysVal(d interface{}) ([]byte, error) {
	if s, ok := d.(string); ok {
		if !utf8.ValidString(s) {
			return nil, fmt.Errorf("'%s' is not a valid UTF-8 string", d)
		}
		return []byte(s), nil
	}
	return nil, fmt.Errorf("Not a string value!")
}

// GetSysValFromString returns the []byte representation of a given string.
func (s *StringAttribute) GetSysValFromString(str string) []byte {
	if !utf8.ValidString(str) {
		panic(fmt.Errorf("'%s' is not a valid UTF-8 coded string!"))
	}
	return []byte(str)
}

// GetStringFromSysVal returns the string representation of a []byte.
func (s *StringAttribute) GetStringFromSysVal(d []byte) string {
	ret := string(d)
	if !utf8.ValidString(ret) {
		panic(fmt.Errorf("'%s' is not a valid UTF-8 coded string!"))
	}
	return ret
}

// Equals checks for equality against another Attribute.
func (s *StringAttribute) Equals(other Attribute) bool {
	if o, ok := other.(*StringAttribute); ok {
		return s.Name == o.Name
	}
	return false
}

// Compatible checks for equality against another Attribute.
func (s *StringAttribute) Compatible(other Attribute) bool {
	if _, ok := other.(*StringAttribute); ok {
		return true
	}
	return false
}

// GetLengthFromSysVal returns the amount of storage the given system value needs.
func (s *StringAttribute) GetLengthFromSysVal(d []byte) int {
	return len(d)
}
