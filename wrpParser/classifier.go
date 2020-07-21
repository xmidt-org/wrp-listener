package wrpparser

import (
	"errors"
	"regexp"

	"github.com/xmidt-org/wrp-go/v2"
)

var (
	errNilRegex         = errors.New("invalid regex, cannot be nil")
	errEmptyClassifiers = errors.New("not enough valid classifiers, must have at least 1")
)

// Classifier is an object that Labels a wrp message.  The Classifier provides
// the label associated with the wrp upon labelling it, as well as a boolean
// that describes whether or not the Classifier could label the message.  The
// message is not modified in any way.
type Classifier interface {
	Label(*wrp.Message) (string, bool)
}

// ConstClassifier returns the same label and ok value for every message the
// struct is asked to label.
type ConstClassifier struct {
	label string
	ok    bool
}

// Label returns the saved label and boolean value without any checks to the
// message it receives.
func (c ConstClassifier) Label(_ *wrp.Message) (string, bool) {
	return c.label, c.ok
}

// NewConstClassifier creates the classifier that consistenly labels every wrp
// message the same way.
func NewConstClassifier(label string, ok bool) *ConstClassifier {
	return &ConstClassifier{label: label, ok: ok}
}

// RegexpLabel labels wrp messages if a regular expression matches a specified
// field.
type RegexpLabel struct {
	label string
	regex *regexp.Regexp
	field Field
}

// Label checks a field in the message given and if the regular expression
// matches against it, returns the stored label.  Otherwise, it returns an
// empty string.  If a label is returned, the boolean is true.  Otherwise,
// false.
func (r *RegexpLabel) Label(msg *wrp.Message) (string, bool) {
	loc := r.regex.FindStringIndex(getFieldValue(r.field, msg))
	if loc == nil {
		return "", false
	}
	return r.label, true
}

// NewRegexpLabel creates a new RegexpLabel struct as long as the regular
// expression provided is valid.
func NewRegexpLabel(label string, regex *regexp.Regexp, field Field) (*RegexpLabel, error) {
	if regex == nil {
		return nil, errNilRegex
	}
	return &RegexpLabel{
		label: label,
		regex: regex,
		field: field,
	}, nil
}

// NewRegexpLabelFromStr takes a string representation of a regular expression
// and compiles it, then creates a RegexpLabel struct.
func NewRegexpLabelFromStr(label, regexStr string, field Field) (*RegexpLabel, error) {
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, err
	}
	return NewRegexpLabel(label, regex, field)
}

// MultClassifier runs multiple classifiers on each message it receives,
// returning the first label that is successfully applied to the message.
type MultClassifier struct {
	classifiers []Classifier
}

// Label attempts to label the wrp message with each Classifier in the
// MultClassifier's list, returning the first successful label.  If no
// Classifier labels the message, an empty string and false boolean is returned.
func (m *MultClassifier) Label(msg *wrp.Message) (string, bool) {
	for _, c := range m.classifiers {
		l, ok := c.Label(msg)
		if ok {
			return l, true
		}
	}
	return "", false
}

// NewMultClassifier builds the MultClassifier struct.  It requires at least
// one valid Classifier.
func NewMultClassifier(classifiers ...Classifier) (*MultClassifier, error) {
	m := MultClassifier{
		classifiers: []Classifier{},
	}
	for _, c := range classifiers {
		// don't include nil classifiers
		if c == nil {
			continue
		}
		m.classifiers = append(m.classifiers, c)
	}
	if len(m.classifiers) == 0 {
		return nil, errEmptyClassifiers
	}
	return &m, nil
}
