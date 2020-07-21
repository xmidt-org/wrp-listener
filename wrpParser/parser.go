package wrpparser

import (
	"errors"

	"github.com/xmidt-org/wrp-go/v2"
)

var (
	errNilFinder     = errors.New("invalid default finder: cannot be nil")
	errNilClassifier = errors.New("invalid classifier: cannot be nil")
)

// DeviceFinder extracts a device id from a wrp message.
type DeviceFinder interface {
	FindDeviceID(msg *wrp.Message) (string, error)
}

// StrParser finds the label for a wrp message and then uses the device finder
// associated with that label to get the device ID associated with the message.
type StrParser struct {
	classifier    Classifier
	finders       map[string]DeviceFinder
	defaultFinder DeviceFinder
}

// Parse takes a message and parses the device ID from it.  It uses the
// Classifier to determine the label associated with the wrp message.  The
// device ID is found using the device finder associated with the message's
// label.  If there is a problem, the default DeviceFinder is used.
func (p *StrParser) Parse(msg *wrp.Message) (string, error) {
	f := p.defaultFinder
	label, ok := p.classifier.Label(msg)

	// if we labelled the message, get the associated DeviceFinder.
	if ok {
		if f, ok = p.finders[label]; !ok {
			// if we don't have a DeviceFinder for the label, use the default.
			f = p.defaultFinder
		}
	}

	return f.FindDeviceID(msg)
}

// ParserOption is a function used to configure the StrParser.
type ParserOption func(*StrParser)

// WithDeviceFinder adds a DeviceFinder that the StrParser will use to find the
// device id of a wrp message.  Each DeviceFinder is associated with a label.
// If the label already has a DeviceFinder associated with it, it will be
// replaced by the new one (as long as the DeviceFinder is not nil).
func WithDeviceFinder(label string, finder DeviceFinder) ParserOption {
	return func(parser *StrParser) {
		if finder != nil {
			parser.finders[label] = finder
		}
	}
}

// NewStrParser sets up the StrParser with a valid classifier and defaultFinder.
// Options need to be provided in order to add more DeviceFinders for diferent
// labels.
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
