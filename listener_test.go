// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/webhook-schema"
)

type vador func(*assert.Assertions, *Listener)

type newTest struct {
	description    string
	r              webhook.Registration
	noRegistration bool
	noUrl          bool
	opt            Option
	opts           []Option
	check          vador
	checks         []vador
	expectedErr    error
}

func vadorBody(assert *assert.Assertions, l *Listener) {
	assert.NotNil(l.body)
}

func vadorAcceptedSecrets(ok ...string) vador {
	return func(assert *assert.Assertions, l *Listener) {
		if ok == nil {
			ok = []string{}
		}
		assert.Equal(ok, l.acceptedSecrets)
	}
}

func vadorGetAuth(want string, err ...error) vador {
	return func(assert *assert.Assertions, l *Listener) {
		got, e := l.getAuth()
		assert.Equal(want, got)

		if err != nil {
			assert.ErrorIs(e, err[0])
		}
	}
}

var validWHR = webhook.Registration{
	Duration: webhook.CustomDuration(5 * time.Minute),
}

// TestNew is focused on validating the input to New and not all the forms of
// options.
func TestNew(t *testing.T) {
	tests := []newTest{
		{
			description: "empty is not ok",
			expectedErr: ErrInput,
		}, {
			description: "no url fails",
			r:           validWHR,
			noUrl:       true,
			expectedErr: ErrInput,
		}, {
			description:    "nil registration fails",
			noRegistration: true,
			expectedErr:    ErrInput,
		}, {
			description: "nearly empty is ok",
			r:           validWHR,
			checks: []vador{
				vadorBody,
				vadorAcceptedSecrets(),
				vadorGetAuth(""),
			},
		}, {
			description: "nearly empty with an interval is ok",
			opt:         Interval(5 * time.Minute),
			r:           validWHR,
			checks: []vador{
				vadorBody,
				vadorAcceptedSecrets(),
				vadorGetAuth(""),
			},
		}, {
			description: "nearly empty with an invalid interval",
			opt:         Interval(-5 * time.Minute),
			expectedErr: ErrInput,
		},
	}
	commonNewTest(t, tests)
}

func commonNewTest(t *testing.T, tests []newTest) {
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			r := tc.r
			opts := make([]Option, 0, len(tc.opts)+1)
			opts = append(opts, tc.opt)
			opts = append(opts, tc.opts...)
			url := "http://example.com"
			if tc.noUrl {
				url = ""
			}
			rPtr := &r
			if tc.noRegistration {
				rPtr = nil
			}
			got, err := New(rPtr, url, opts...)

			if tc.expectedErr != nil {
				assert.Nil(got)
				assert.ErrorIs(err, tc.expectedErr)
				return
			}

			require.NotNil(got)
			require.NoError(err)

			checks := make([]vador, 0, len(tc.checks)+1)
			checks = append(checks, tc.check)
			checks = append(checks, tc.checks...)
			for _, c := range checks {
				if c != nil {
					c(assert, got)
				}
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		description string
		input       http.Request
		opt         Option
		opts        []Option
		expected    Token
		expectedErr error
	}{
		{
			description: "basic test",
			input: http.Request{
				Header: http.Header{
					webpaHeader: []string{"sha1=12345"},
				},
			},
			opt: AcceptSHA1(),
			expected: token{
				alg:       "sha1",
				principal: "12345",
			},
		}, {
			description: "basic test, using the alternate header",
			input: http.Request{
				Header: http.Header{
					xmidtHeader: []string{"sha1=12345"},
				},
			},
			opt: AcceptSHA1(),
			expected: token{
				alg:       "sha1",
				principal: "12345",
			},
		}, {
			description: "multiple auth possible, choose the best",
			input: http.Request{
				Header: http.Header{
					webpaHeader: []string{"sha1=12345"},
				},
			},
			opts: []Option{
				AcceptSHA1(),
				AcceptNoHash(),
			},
			expected: token{
				alg:       "sha1",
				principal: "12345",
			},
		}, {
			description: "no header with that name",
			opt:         AcceptNoHash(),
			expected: token{
				alg:       "none",
				principal: "",
			},
		}, {
			description: "empty header",
			input: http.Request{
				Header: http.Header{
					webpaHeader: []string{"   "},
				},
			},
			opt: AcceptNoHash(),
			expected: token{
				alg:       "none",
				principal: "",
			},
		}, {
			description: "malformed header",
			input: http.Request{
				Header: http.Header{
					webpaHeader: []string{"foo=="},
				},
			},
			expectedErr: ErrInvalidAuth,
		}, {
			description: "malformed header 2",
			input: http.Request{
				Header: http.Header{
					webpaHeader: []string{"foo"},
				},
			},
			expectedErr: ErrInvalidAuth,
		}, {
			description: "empty value",
			input: http.Request{
				Header: http.Header{
					webpaHeader: []string{"foo=  "},
				},
			},
			expectedErr: ErrInvalidAuth,
		}, {
			description: "empty key",
			input: http.Request{
				Header: http.Header{
					webpaHeader: []string{"=foo"},
				},
			},
			expectedErr: ErrInvalidAuth,
		}, {
			description: "no matching key",
			input: http.Request{
				Header: http.Header{
					webpaHeader: []string{"sha256=12345"},
				},
			},
			expectedErr: ErrNotAcceptedHash,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			opts := append(tc.opts, tc.opt)
			whl, err := New(&webhook.Registration{
				Duration: webhook.CustomDuration(5 * time.Minute),
			},
				"http://example.com",
				opts...,
			)

			require.NotNil(whl)
			require.NoError(err)

			in := tc.input
			got, err := whl.Tokenize(&in)

			if tc.expectedErr != nil {
				assert.Nil(got)
				assert.ErrorIs(err, tc.expectedErr)
				return
			}

			assert.NoError(err)
			require.NotNil(got)
			assert.Equal(tc.expected.Type(), got.Type())
			assert.Equal(tc.expected.Principal(), got.Principal())
		})
	}
}

func TestAuthorize(t *testing.T) {
	tests := []struct {
		description string
		input       http.Request
		token       Token
		opt         Option
		opts        []Option
		expectedErr error
	}{
		{
			description: "basic test",
			input: http.Request{
				Body: io.NopCloser(strings.NewReader("foo")),
			},
			opts: []Option{
				AcceptSHA1(),
				AcceptedSecrets("123456"),
			},
			token: token{
				alg:       "sha1",
				principal: "f76a55b14b2b3bd08116b4ee857dd6439b507317",
			},
		}, {
			description: "basic test, signature does not match",
			input: http.Request{
				Body: io.NopCloser(strings.NewReader("foo")),
			},
			opts: []Option{
				AcceptSHA1(),
				AcceptedSecrets("123456"),
			},
			token: token{
				alg:       "sha1",
				principal: "0000",
			},
			expectedErr: ErrInput,
		}, {
			description: "empty body",
			input: http.Request{
				Body: io.NopCloser(strings.NewReader("")),
			},
			opts: []Option{
				AcceptSHA1(),
				AcceptedSecrets("123456"),
			},
			expectedErr: ErrInput,
			token: token{
				alg:       "sha1",
				principal: "0000",
			},
		}, {
			description: "no body",
			opts: []Option{
				AcceptSHA1(),
				AcceptedSecrets("123456"),
			},
			expectedErr: ErrInput,
			token: token{
				alg:       "sha1",
				principal: "0000",
			},
		}, {
			description: "invalid principle",
			token: token{
				alg:       "sha1",
				principal: "f", // invalid because it needs to be 2 characters.
			},
			expectedErr: ErrInput,
		}, {
			description: "nil token",
			expectedErr: ErrInput,
		}, {
			description: "no matching hash",
			input: http.Request{
				Body: io.NopCloser(strings.NewReader("foo")),
			},
			token: token{
				alg:       "sha1",
				principal: "f0",
			},
			expectedErr: ErrNotAcceptedHash,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			opts := append(tc.opts, tc.opt)
			whl, err := New(
				&webhook.Registration{
					Duration: webhook.CustomDuration(5 * time.Minute),
				},
				"http://example.com",
				opts...,
			)

			require.NotNil(whl)
			require.NoError(err)

			in := tc.input

			err = whl.Authorize(&in, tc.token)

			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
				return
			}

			assert.NoError(err)
		})
	}
}

func TestListener_Accept(t *testing.T) {
	tests := []struct {
		description  string
		opt          Option
		opts         []Option
		expectBefore []string
		secrets      []string
	}{
		{
			description: "simple test",
			secrets:     []string{"foo"},
		}, {
			description:  "simple test accepting secrets",
			opt:          AcceptedSecrets("bar"),
			expectBefore: []string{"bar"},
			secrets:      []string{"foo"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			r := validWHR
			opts := append(tc.opts, tc.opt)
			l, err := New(&r, "http://example.com", opts...)
			require.NotNil(l)
			require.NoError(err)

			if tc.expectBefore == nil {
				tc.expectBefore = []string{}
			}
			assert.Equal(tc.expectBefore, l.acceptedSecrets)

			l.Accept(tc.secrets)
			assert.Equal(tc.secrets, l.acceptedSecrets)
		})
	}
}

func TestListener_String(t *testing.T) {
	tests := []struct {
		description string
		opt         Option
		opts        []Option
		str         string
	}{
		{
			description: "simple test",
			str:         "Listener(URL(http://example.com))",
		}, {
			description: "simple test",
			opt:         AcceptedSecrets("bar"),
			str:         "Listener(URL(http://example.com), AcceptedSecrets(***))",
		}, {
			description: "simple test",
			opts:        []Option{AcceptedSecrets("bar"), AcceptSHA1()},
			str:         "Listener(URL(http://example.com), AcceptedSecrets(***), AcceptSHA1())",
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			r := validWHR
			opts := append(tc.opts, tc.opt)
			l, err := New(&r, "http://example.com", opts...)
			require.NotNil(l)
			require.NoError(err)

			assert.Equal(tc.str, l.String())
		})
	}
}
