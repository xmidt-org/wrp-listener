// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	tests := []struct {
		description string
		reg         *Registration
		token       *Tokenize
		auth        *Authorize
		want        string
	}{
		{
			description: "Empty Registration",
			reg:         &Registration{},
			want: "event.Registration{\n" +
				"  At:         0001-01-01T00:00:00Z\n" +
				"  Duration:   0s\n" +
				"  Body:       ''\n" +
				"  StatusCode: 0\n" +
				"  Until:      0001-01-01T00:00:00Z\n" +
				"  Err:        <nil>\n" +
				"}\n",
		}, {
			description: "Empty Tokenize",
			token:       &Tokenize{},
			want: "event.Tokenize{\n" +
				"  Header:     ''\n" +
				"  Algorithms: []\n" +
				"  Algorithm:  ''\n" +
				"  Err:        <nil>\n" +
				"}\n",
		}, {
			description: "Empty Authorize",
			auth:        &Authorize{},
			want: "event.Authorize{\n" +
				"  Algorithm:  ''\n" +
				"  Err:        <nil>\n" +
				"}\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			switch {
			case tc.reg != nil:
				assert.Equal(tc.want, tc.reg.String())
			case tc.token != nil:
				assert.Equal(tc.want, tc.token.String())
			case tc.auth != nil:
				assert.Equal(tc.want, tc.auth.String())
			}
		})
	}
}

func TestRegistrationListenerFunc(t *testing.T) {
	assert := assert.New(t)

	var called bool
	f := RegistrationFunc(func(Registration) {
		called = true
	})

	f.OnRegistrationEvent(Registration{})
	assert.True(called)
}

func TestTokenizeListenerFunc(t *testing.T) {
	assert := assert.New(t)

	var called bool
	f := TokenizeFunc(func(Tokenize) {
		called = true
	})

	f.OnTokenizeEvent(Tokenize{})
	assert.True(called)
}

func TestAuthorizeListenerFunc(t *testing.T) {
	assert := assert.New(t)

	var called bool
	f := AuthorizeFunc(func(Authorize) {
		called = true
	})

	f.OnAuthorizeEvent(Authorize{})
	assert.True(called)
}
