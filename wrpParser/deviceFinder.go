package wrpparser

import (
	"errors"
	"regexp"
	"strings"

	"github.com/xmidt-org/wrp-go/v2"
)

type FieldFinder struct {
	Field field
}

func (f FieldFinder) FindDeviceID(msg *wrp.Message) (string, error) {
	return strings.ToLower(getFieldValue(f.Field, msg)), nil
}

type RegexpFinder struct {
	field            field
	regex            *regexp.Regexp
	deviceLabel      string
	subExpressionIdx int
}

func (r *RegexpFinder) FindDeviceID(msg *wrp.Message) (string, error) {
	fieldValue := getFieldValue(r.field, msg)
	matches := r.regex.FindStringSubmatch(fieldValue)
	if matches == nil || len(matches) == 0 || len(matches) <= r.subExpressionIdx {
		return "", errors.New("some error")
	}
	deviceID := matches[r.subExpressionIdx]
	if deviceID == "" {
		return "", errors.New("some error")
	}
	return strings.ToLower(deviceID), nil
}

func NewRegexpFinder(field field, regex *regexp.Regexp, deviceLabel string) (*RegexpFinder, error) {
	if regex == nil {
		return nil, errors.New("some error")
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
	return nil, errors.New("some error")
}

func NewRegexpFinderFromStr(field field, regexStr string, deviceLabel string) (*RegexpFinder, error) {
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, err
	}
	return NewRegexpFinder(field, regex, deviceLabel)
}
