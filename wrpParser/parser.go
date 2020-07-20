package wrpparser

import (
	"errors"

	"github.com/xmidt-org/wrp-go/v2"
)

var (
	errNilFinder     = errors.New("invalid default finder: cannot be nil")
	errNilClassifier = errors.New("invalid classifier: cannot be nil")
)

type DeviceFinder interface {
	FindDeviceID(msg *wrp.Message) (string, error)
}

type Classifier interface {
	Label(*wrp.Message) (string, bool)
}

type StrParser struct {
	classifier    Classifier
	finders       map[string]DeviceFinder
	defaultFinder DeviceFinder
}

func (p *StrParser) Parse(msg *wrp.Message) (string, error) {
	f := p.defaultFinder
	label, ok := p.classifier.Label(msg)

	// if we labelled the message, get the associated DeviceFinder
	if ok {
		f, ok = p.finders[label]
		// if we don't have a DeviceFinder for the label, use the default
		if !ok {
			f = p.defaultFinder
		}
	}

	return f.FindDeviceID(msg)

}

type ParserOption func(*StrParser)

func WithDeviceFinder(label string, finder DeviceFinder) ParserOption {
	return func(parser *StrParser) {
		if finder != nil {
			parser.finders[label] = finder
		}
	}
}

func NewStrParser(classifier Classifier, defaultFinder DeviceFinder, options ...ParserOption) (*StrParser, error) {
	if defaultFinder == nil {
		return nil, errNilFinder
	}

	if classifier == nil {
		return nil, errNilClassifier
	}

	p := &StrParser{
		defaultFinder: defaultFinder,
		classifier:    classifier,
		finders:       make(map[string]DeviceFinder),
	}

	for _, o := range options {
		o(p)
	}

	return p, nil
}
