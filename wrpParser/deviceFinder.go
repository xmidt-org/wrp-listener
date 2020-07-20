package wrpparser

import (
	"errors"
	"regexp"
	"strings"

	"github.com/xmidt-org/wrp-go/v2"
)

var (
	errNoMatch            = errors.New("regular expression didn't match field")
	errEmptyDeviceID      = errors.New("device id is empty string")
	errRegexLabelMismatch = errors.New("regular expression doesn't have the label provided for finding the device id")
)

type FieldFinder struct {
	Field Field
}

func (f FieldFinder) FindDeviceID(msg *wrp.Message) (string, error) {
	return strings.ToLower(getFieldValue(f.Field, msg)), nil
}

type RegexpFinder struct {
	field            Field
	regex            *regexp.Regexp
	deviceLabel      string
	subExpressionIdx int
}

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

func NewRegexpFinderFromStr(field Field, regexStr string, deviceLabel string) (*RegexpFinder, error) {
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, err
	}
	return NewRegexpFinder(field, regex, deviceLabel)
}
