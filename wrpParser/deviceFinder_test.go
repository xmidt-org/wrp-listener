package wrpparser

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/wrp-go/v2"
)

func TestFieldFinder(t *testing.T) {
	tests := []struct {
		description      string
		msg              *wrp.Message
		field            Field
		expectedDeviceID string
	}{
		{
			description: "destination",
			msg: &wrp.Message{
				Destination: "DESTINATION",
			},
			field:            Destination,
			expectedDeviceID: "destination",
		},
		{
			description: "source",
			msg: &wrp.Message{
				Source: "mEh",
			},
			field:            Source,
			expectedDeviceID: "meh",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			finder := FieldFinder{
				Field: tc.field,
			}
			result, err := finder.FindDeviceID(tc.msg)
			assert.Nil(err)
			assert.Equal(tc.expectedDeviceID, result)
		})
	}
}

func TestNewRegexpFinder(t *testing.T) {
	goodRegexp, err := regexp.Compile("(?P<device>.*)")
	assert.Nil(t, err)
	tests := []struct {
		description    string
		regex          *regexp.Regexp
		label          string
		finderReturned bool
		expectedErr    error
	}{
		{
			description:    "Success",
			regex:          goodRegexp,
			label:          "device",
			finderReturned: true,
		},
		{
			description: "Nil Regex Error",
			regex:       nil,
			label:       "",
			expectedErr: errNilRegex,
		},
		{
			description: "Regex Label Match Error",
			regex:       goodRegexp,
			label:       "bad",
			expectedErr: errRegexLabelMismatch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			finder, err := NewRegexpFinder(Source, tc.regex, tc.label)
			if tc.finderReturned {
				assert.NotNil(finder)
			} else {
				assert.Nil(finder)
			}
			assert.Equal(tc.expectedErr, err)
		})
	}
}

func TestNewRegexpFinderFromStr(t *testing.T) {
	tests := []struct {
		description    string
		regex          string
		finderReturned bool
		expectedErr    error
	}{
		{
			description:    "Success",
			regex:          "(?P<device>.*)",
			finderReturned: true,
		},
		{
			description: "Regexp Compile Error",
			regex:       `\V`,
			expectedErr: errors.New("failed to compile"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			finder, err := NewRegexpFinderFromStr(Destination, tc.regex, "device")
			if tc.finderReturned {
				assert.NotNil(finder)
			} else {
				assert.Nil(finder)
			}
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestRegexpFinder(t *testing.T) {
	finder, err := NewRegexpFinderFromStr(Destination, "device/(?P<device>.*)/[A-Za-z]+", "device")
	assert.Nil(t, err)
	tests := []struct {
		description    string
		msg            *wrp.Message
		expectedResult string
		expectedErr    error
	}{
		{
			description: "Success",
			msg: &wrp.Message{
				Destination: "device/mac:Whatever/abcd",
			},
			expectedResult: "mac:whatever",
		},
		{
			description: "No Match Error",
			msg: &wrp.Message{
				Destination: "device/mac:whatever/",
			},
			expectedErr: errNoMatch,
		},
		{
			description: "Empty Device ID Error",
			msg: &wrp.Message{
				Destination: "device//mac:whatever",
			},
			expectedErr: errEmptyDeviceID,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			result, err := finder.FindDeviceID(tc.msg)
			assert.Equal(tc.expectedResult, result)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}
