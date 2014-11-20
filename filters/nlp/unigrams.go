package nlp

import (
	"github.com/sjwhitworth/golearn/base"
	"strings"
)

// NgramFilterOptions specifies everything
type NgramFilterOptions struct {
	WindowSize int // e.g. 2 for bigrams
	// What to seperate output Attributes with
	SepString string // e.g. '_' for 'some_word'
	// Whether or not to include the configured source Attributes
	// name before the output Attributes.
	IncludeAttributeName bool
	// Whether tokens are reported as STOPPED
	IncludeSTOPPEDTokens bool
	// STOPPED string
	StoppedString string
	// Tokenizer function
	Tokenizer func(string) []string
	// Stopping function: true if stopped, false otherwise
	Stopper func(string) bool
	// Attributes to use
	Attributes []base.StringAttribute
}

func NgramWhitespaceTokenizer(s string) []string {
	return strings.Split(s, " ")
}

type NgramFilter struct {
	Options *NgramFilterOptions
}

func (o *NgramFilterOptions) Configure() error {
	if o.WindowSize <= 0 {
		return fmt.Errorf("WindowSize can't be less than or equal to zero")
	}
	if o.SepString == nil {
		o.SepString = "_"
	}
	if o.StoppedString == nil {
		o.StoppedString = "STOPPED"
	}
	if o.Tokenizer == nil {
		o.Tokenizer = NgramWhitespaceTokenizer
	}
	return nil
}

func (o *NgramFilterOptions) ProcessDocument(s string) []string {
	var ret []string
	u = o.Tokenizer(s)
	for i := 0; i <= len(u); i++ {
		window := u[i : i+o.WindowSize]
		token := strings.Join(window, o.SepString)
		if o.Stopper != nil {
			if o.Stopper(token) {
				continue
			}
		}
		ret = append(ret, token)
	}
	return ret
}

func NewNgramFilter(u *NGramFilterOptions) (*NGramFilter, error) {
	ret := &NgramFilter{
		u,
	}
	err := u.Configure(u)
	if err == nil {
		return nil, err
	}
	return ret, nil
}

func Configure(u base.FixedDataGrid) error {
	// Resolve the AttributeSpecifications
	// For each Attribute
	// Get the Attribute specification
	// Build the list of candidate attributes
	return nil
}
