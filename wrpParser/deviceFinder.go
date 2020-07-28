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
	"strings"

	"github.com/xmidt-org/wrp-go/v3"
)

var (
	errNoMatch            = errors.New("regular expression didn't match field")
	errEmptyDeviceID      = errors.New("device id is empty string")
	errRegexLabelMismatch = errors.New("regular expression doesn't have the label provided for finding the device id")
)

// TODO: should we validate the device id against the schema expected?
// https://xmidt.io/docs/wrp/basics/#device-identification

// FieldFinder returns the full value of the given field in a wrp message.
type FieldFinder struct {
	Field Field
}

// FindDeviceID expects to find the device ID as the only value in a specific
// field in the wrp message.  It returns a lowercase transform of this.
func (f FieldFinder) FindDeviceID(msg *wrp.Message) (string, error) {
	return strings.ToLower(getFieldValue(f.Field, msg)), nil
}

// RegexpFinder uses a regular expression to find the device id within a field
// of a wrp message.
type RegexpFinder struct {
	field            Field
	regex            *regexp.Regexp
	deviceLabel      string
	subExpressionIdx int
}

// FindDeviceID applies a regular expression to a specific field in the wrp
// message.  If that is successful, it extracts the device id from the expected
// place in the regular expression submatch.
func (r *RegexpFinder) FindDeviceID(msg *wrp.Message) (string, error) {
	fieldValue := getFieldValue(r.field, msg)
	matches := r.regex.FindStringSubmatch(fieldValue)
	if matches == nil || len(matches) == 0 || len(matches) <= r.subExpressionIdx {
		return "", errNoMatch
	}
	deviceID := matches[r.subExpressionIdx]
	if deviceID == "" {
		return "", errEmptyDeviceID
	}
	return strings.ToLower(deviceID), nil
}

// NewRegexpFinder returns a new RegexpFinder that checks the field given using
// the regular expression provided.  It will extract the device id from the
// regular expression result at the label given.
func NewRegexpFinder(field Field, regex *regexp.Regexp, deviceLabel string) (*RegexpFinder, error) {
	if regex == nil {
		return nil, errNilRegex
	}
	r := RegexpFinder{
		field:       field,
		regex:       regex,
		deviceLabel: deviceLabel,
	}
	for idx, val := range r.regex.SubexpNames() {
		if val == r.deviceLabel {
			r.subExpressionIdx = idx
			return &r, nil
		}
	}
	return nil, errRegexLabelMismatch
}

// NewRegexpFinderFromStr compiles a string representation of a regular
// expression and returns a new RegexpFinder that checks the field given using
// that regular expression.  It will extract the device id from the regular
// expression result at the label given.
func NewRegexpFinderFromStr(field Field, regexStr string, deviceLabel string) (*RegexpFinder, error) {
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile [%v] into a regular expression: %w", regexStr, err)
	}
	return NewRegexpFinder(field, regex, deviceLabel)
}
