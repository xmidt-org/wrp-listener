// SPDX-FileCopyrightText: 2020 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpparser

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/wrp-go/v3"
)

func TestConstClassifier(t *testing.T) {
	assert := assert.New(t)
	var msg *wrp.Message
	label := "testLabel"
	neverLabel := NewConstClassifier("", false)
	alwaysLabel := NewConstClassifier(label, true)

	result, ok := neverLabel.Label(msg)
	assert.Empty(result)
	assert.False(ok)

	result, ok = alwaysLabel.Label(msg)
	assert.Equal(label, result)
	assert.True(ok)

}

func TestConstClassifierInterface(t *testing.T) {
	// test that the ConstClassifier implements the Classifier interface.
	var classifier interface{} = NewConstClassifier("", false)
	_, ok := classifier.(Classifier)
	assert.True(t, ok)
}

func TestNewRegexpClassifier(t *testing.T) {
	assert := assert.New(t)

	// success
	regex, err := regexp.Compile("")
	assert.Nil(err)
	c, err := NewRegexpClassifier("", regex, Source)
	assert.NotNil(c)
	assert.Nil(err)

	// error case
	c, err = NewRegexpClassifier("", nil, Source)
	assert.Nil(c)
	assert.Equal(errNilRegex, err)
}

func TestNewRegexpClassifierFromStr(t *testing.T) {
	assert := assert.New(t)

	// success
	c, err := NewRegexpClassifierFromStr("", "", Destination)
	assert.NotNil(c)
	assert.Nil(err)

	// error case
	c, err = NewRegexpClassifierFromStr("", `\V`, Destination)
	assert.Nil(c)
	assert.Contains(err.Error(), "failed to compile")
}

func TestRegexpClassifierInterface(t *testing.T) {
	// test that the RegexpClassifier implements the Classifier interface.
	var (
		classifier interface{}
		err        error
	)
	classifier, err = NewRegexpClassifierFromStr("", "", Destination)
	assert := assert.New(t)
	assert.Nil(err)
	_, ok := classifier.(Classifier)
	assert.True(ok)
}

func TestRegexpClassifierLabel(t *testing.T) {
	classifier, err := NewRegexpClassifierFromStr("event", "[0-9]+", Source)
	assert.NotNil(t, classifier)
	assert.Nil(t, err)

	tests := []struct {
		description   string
		msg           *wrp.Message
		expectedLabel string
		expectedBool  bool
	}{
		{
			description:   "Success",
			msg:           &wrp.Message{Source: "4214184918249712"},
			expectedLabel: "event",
			expectedBool:  true,
		},
		{
			description:  "Nil Message Error",
			expectedBool: false,
		},
		{
			description:  "No Match Error",
			msg:          &wrp.Message{Source: "aaaaaaa"},
			expectedBool: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			label, ok := classifier.Label(tc.msg)
			assert.Equal(tc.expectedBool, ok)
			assert.Equal(tc.expectedLabel, label)
		})
	}
}

func TestNewMultClassifierError(t *testing.T) {
	var nilC Classifier
	assert := assert.New(t)
	c, err := NewMultClassifier(nilC, nilC, nilC, nilC)
	assert.Nil(c)
	assert.Equal(errEmptyClassifiers, err)
}

func TestMultClassifier(t *testing.T) {
	// successfully create a MultClassifier
	var c *MultClassifier
	t.Run("TestNewMultClassifierSuccess", func(t *testing.T) {
		require := require.New(t)
		c = newMultClassifierSuccess(require)
	})

	// don't move forward with other tests if the classifier isn't good
	require.NotNil(t, c)

	// ensure it implements a Classifier interface
	t.Run("TestRegexpClassifierInterface", func(t *testing.T) {
		isMultClassifierInterface(require.New(t), c)
	})

	// test Label()
	t.Run("TestMultClassifierLabel", func(t *testing.T) {
		testMultClassifierLabel(t, c)
	})
}

func newMultClassifierSuccess(a *require.Assertions) *MultClassifier {
	c1, err := NewRegexpClassifierFromStr("event", "[0-9]+", Source)
	a.Nil(err)
	a.NotNil(c1)

	c2, err := NewRegexpClassifierFromStr("something else", ".+", Destination)
	a.Nil(err)
	a.NotNil(c2)

	classifier, err := NewMultClassifier(c1, c2)
	a.Nil(err)
	a.NotNil(classifier)

	return classifier
}

func isMultClassifierInterface(a *require.Assertions, classifier interface{}) {
	// test that the MultClassifier implements the Classifier interface.
	_, ok := classifier.(Classifier)
	a.True(ok)
}

func testMultClassifierLabel(t *testing.T, classifier *MultClassifier) {
	tests := []struct {
		description   string
		msg           *wrp.Message
		expectedLabel string
		expectedBool  bool
	}{
		{
			description:   "Success",
			msg:           &wrp.Message{Destination: "somewhere"},
			expectedLabel: "something else",
			expectedBool:  true,
		},
		{
			description: "Multi Label Success",
			msg: &wrp.Message{
				Source:      "9999",
				Destination: "yay",
			},
			expectedLabel: "event",
			expectedBool:  true,
		},
		{
			description:   "No Match Error",
			msg:           &wrp.Message{},
			expectedLabel: "",
			expectedBool:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			label, ok := classifier.Label(tc.msg)
			assert.Equal(tc.expectedLabel, label)
			assert.Equal(tc.expectedBool, ok)
		})
	}
}
