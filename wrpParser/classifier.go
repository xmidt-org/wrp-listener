/**
 * Copyright 2020 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package wrpparser

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/xmidt-org/wrp-go/v3"
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
// message it receives.  Even if the message is nil, the same result is
// provided.
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
type RegexpClassifier struct {
	label string
	regex *regexp.Regexp
	field Field
}

// Label checks a field in the message given and if the regular expression
// matches against it, returns the stored label.  Otherwise, it returns an
// empty string.  If a label is returned, the boolean is true.  Otherwise,
// false.
func (r *RegexpClassifier) Label(msg *wrp.Message) (string, bool) {
	// if the wrp message is nil, don't do anything
	if msg == nil {
		return "", false
	}

	loc := r.regex.FindStringIndex(getFieldValue(r.field, msg))
	if loc == nil {
		return "", false
	}
	return r.label, true
}

// NewRegexpClassifier creates a new RegexpClassifier struct as long as the regular
// expression provided is valid.
func NewRegexpClassifier(label string, regex *regexp.Regexp, field Field) (*RegexpClassifier, error) {
	if regex == nil {
		return nil, errNilRegex
	}
	return &RegexpClassifier{
		label: label,
		regex: regex,
		field: field,
	}, nil
}

// NewRegexpClassifierFromStr takes a string representation of a regular expression
// and compiles it, then creates a RegexpClassifier struct.
func NewRegexpClassifierFromStr(label, regexStr string, field Field) (*RegexpClassifier, error) {
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile [%v] into a regular expression: %w", regexStr, err)
	}
	return NewRegexpClassifier(label, regex, field)
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
