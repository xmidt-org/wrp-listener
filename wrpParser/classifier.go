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

type ConstClassifier struct {
	label string
	ok    bool
}

func (c ConstClassifier) Label(_ *wrp.Message) (string, bool) {
	return c.label, c.ok
}

func NewConstClassifier(label string, ok bool) *ConstClassifier {
	return &ConstClassifier{label: label, ok: ok}
}

type RegexpLabel struct {
	label string
	regex *regexp.Regexp
	field Field
}

func (r *RegexpLabel) Label(msg *wrp.Message) (string, bool) {
	loc := r.regex.FindStringIndex(getFieldValue(r.field, msg))
	if loc == nil {
		return "", false
	}
	return r.label, true
}

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

func NewRegexpLabelFromStr(label, regexStr string, field Field) (*RegexpLabel, error) {
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, err
	}
	return NewRegexpLabel(label, regex, field)
}

type MultClassifier struct {
	classifiers []Classifier
}

func (m *MultClassifier) Label(msg *wrp.Message) (string, bool) {
	for _, c := range m.classifiers {
		l, ok := c.Label(msg)
		if ok {
			return l, true
		}
	}
	return "", false
}

func NewMultClassifier(classifiers ...Classifier) (*MultClassifier, error) {
	m := MultClassifier{
		classifiers: []Classifier{},
	}
	for _, c := range classifiers {
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
