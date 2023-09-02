// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"context"
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/webhook-schema"
	"github.com/xmidt-org/wrp-listener/event"
)

func TestOptionStrings(t *testing.T) {
	var cancel CancelEventListenerFunc

	tests := []struct {
		in       Option
		expected string
	}{
		{
			in:       Interval(5 * time.Minute),
			expected: "Interval(5m0s)",
		}, {
			in:       Once(),
			expected: "Once()",
		}, {
			in:       HTTPClient(http.DefaultClient),
			expected: "HTTPClient(client)",
		}, {
			in:       HTTPClient(nil),
			expected: "HTTPClient(nil)",
		}, {
			in:       DecorateRequest(nil),
			expected: "DecorateRequest(nil)",
		}, {
			in:       DecorateRequest(DecoratorFunc(func(*http.Request) error { return nil })),
			expected: "DecorateRequest(DecoratorFunc(fn))",
		}, {
			in:       AcceptedSecrets("foo"),
			expected: "AcceptedSecrets(***)",
		}, {
			in:       AcceptedSecrets("foo", "bar"),
			expected: "AcceptedSecrets(***, ...)",
		}, {
			in:       AcceptNoHash(),
			expected: "AcceptNoHash()",
		}, {
			in:       AcceptSHA1(),
			expected: "AcceptSHA1()",
		}, {
			in:       AcceptSHA256(),
			expected: "AcceptSHA256()",
		}, {
			in:       AcceptCustom("foo", sha256.New),
			expected: "AcceptCustom(foo, fn)",
		}, {
			in:       WebhookOpts(webhook.AtLeastOneEvent(), webhook.DeviceIDRegexMustCompile()),
			expected: "RegistrationOpts(AtLeastOneEvent(), DeviceIDRegexMustCompile())",
		}, {
			in:       Context(context.Background()),
			expected: "Context(ctx)",
		}, {
			in:       WithAuthorizeEventListener(event.AuthorizeFunc(func(event.Authorize) {}), &cancel),
			expected: "WithAuthorizeEventListener(lstnr, *cancel)",
		}, {
			in:       WithTokenizeEventListener(event.TokenizeFunc(func(event.Tokenize) {}), &cancel),
			expected: "WithTokenizeEventListener(lstnr, *cancel)",
		}, {
			in:       WithRegistrationEventListener(event.RegistrationFunc(func(event.Registration) {}), &cancel),
			expected: "WithRegistrationEventListener(lstnr, *cancel)",
		}, {
			in:       WithAuthorizeEventListener(event.AuthorizeFunc(func(event.Authorize) {})),
			expected: "WithAuthorizeEventListener(lstnr)",
		}, {
			in:       WithTokenizeEventListener(event.TokenizeFunc(func(event.Tokenize) {})),
			expected: "WithTokenizeEventListener(lstnr)",
		}, {
			in:       WithRegistrationEventListener(event.RegistrationFunc(func(event.Registration) {})),
			expected: "WithRegistrationEventListener(lstnr)",
		}, {
			in:       WithAuthorizeEventListener(nil),
			expected: "WithAuthorizeEventListener(nil)",
		}, {
			in:       WithTokenizeEventListener(nil),
			expected: "WithTokenizeEventListener(nil)",
		}, {
			in:       WithRegistrationEventListener(nil),
			expected: "WithRegistrationEventListener(nil)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.in.String())
		})
	}
}

func TestHTTPClient(t *testing.T) {
	client := &http.Client{}

	tests := []newTest{
		{
			description: "assert default client is there",
			r:           validWHR,
			check: func(assert *assert.Assertions, l *Listener) {
				assert.Equal(l.client, http.DefaultClient)
			},
		}, {
			description: "assert new client is there",
			r:           validWHR,
			opt:         HTTPClient(client),
			check: func(assert *assert.Assertions, l *Listener) {
				assert.Equal(l.client, client)
			},
		}, {
			description: "assert nil client works",
			r:           validWHR,
			opts:        []Option{HTTPClient(client), HTTPClient(nil)},
			check: func(assert *assert.Assertions, l *Listener) {
				assert.Equal(l.client, http.DefaultClient)
			},
		},
	}
	commonNewTest(t, tests)
}

func TestSecrets(t *testing.T) {
	tests := []newTest{
		{
			description: "assert default accepted secrets are empty",
			r:           validWHR,
			check:       vadorAcceptedSecrets(),
		}, {
			description: "assert AcceptedSecrets() works",
			r:           validWHR,
			opt:         AcceptedSecrets("foo"),
			check:       vadorAcceptedSecrets("foo"),
		}, {
			description: "assert AcceptedSecrets() works",
			r:           validWHR,
			opt:         AcceptedSecrets("foo", "bar"),
			check:       vadorAcceptedSecrets("foo", "bar"),
		}, {
			description: "assert multiple AcceptedSecrets() works",
			r:           validWHR,
			opts: []Option{
				AcceptedSecrets("foo"),
				AcceptedSecrets("car", "cat"),
				AcceptedSecrets("bar"),
			},
			check: vadorAcceptedSecrets("foo", "car", "cat", "bar"),
		},
	}
	commonNewTest(t, tests)
}

func TestHashes(t *testing.T) {
	tests := []newTest{
		{
			description: "assert default hashes are empty",
			r:           validWHR,
			check: func(assert *assert.Assertions, l *Listener) {
				assert.Empty(l.hashes)
			},
		}, {
			description: "assert none works",
			r:           validWHR,
			opt:         AcceptNoHash(),
			check: func(assert *assert.Assertions, l *Listener) {
				assert.Nil(l.hashes["none"])
			},
		}, {
			description: "assert SHA1 works",
			r:           validWHR,
			opt:         AcceptSHA1(),
			check: func(assert *assert.Assertions, l *Listener) {
				got := l.hashes["sha1"]()
				want := sha1.New() //nolint:gosec
				assert.Equal(want, got)
			},
		}, {
			description: "assert SHA256 works",
			r:           validWHR,
			opt:         AcceptSHA256(),
			check: func(assert *assert.Assertions, l *Listener) {
				got := l.hashes["sha256"]()
				want := sha256.New()
				assert.Equal(want, got)
			},
		}, {
			description: "assert Custom works",
			r:           validWHR,
			opt:         AcceptCustom("foo", sha256.New),
			check: func(assert *assert.Assertions, l *Listener) {
				got := l.hashes["foo"]()
				want := sha256.New()
				assert.Equal(want, got)
			},
		}, {
			description: "assert Custom nil errors",
			r:           validWHR,
			opt:         AcceptCustom("foo", nil),
			expectedErr: ErrInput,
		},
	}
	commonNewTest(t, tests)
}

func TestRegistrationOpts(t *testing.T) {
	tests := []newTest{
		{
			description: "assert RegistrationOpts() works",
			r:           validWHR,
			opt:         WebhookOpts(webhook.DeviceIDRegexMustCompile()),
		}, {
			description: "assert RegistrationOpts() works, catching an error",
			r: webhook.Registration{
				Duration: webhook.CustomDuration(5 * time.Minute),
				Matcher: webhook.MetadataMatcherConfig{
					DeviceID: []string{"invalid \\\\("},
				},
			},
			opt:         WebhookOpts(webhook.DeviceIDRegexMustCompile()),
			expectedErr: ErrInput,
		},
	}
	commonNewTest(t, tests)
}

func TestContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, struct{}{})

	tests := []newTest{
		{
			description: "assert no ctx works",
			r:           validWHR,
			check: func(assert *assert.Assertions, l *Listener) {
				assert.NotEqual(ctx, l.ctx)
				assert.NotEqual(ctx, l.upstreamCtx)
			},
		},
		{
			description: "assert Context() works",
			r:           validWHR,
			opt:         Context(ctx),
			check: func(assert *assert.Assertions, l *Listener) {
				assert.NotEqual(ctx, l.ctx)
				assert.Equal(ctx, l.upstreamCtx)
			},
		},
	}
	commonNewTest(t, tests)
}

func TestWithAuthorizeEventListener(t *testing.T) {
	m := sync.Mutex{}
	var cancel CancelEventListenerFunc
	var cancel2 CancelEventListenerFunc

	var count int
	listener := event.AuthorizeFunc(func(event.Authorize) {
		count++
	})

	assert.Nil(t, cancel)

	tests := []newTest{
		{
			description: "assert WithAuthorizeEventListener() with cancel works",
			r:           validWHR,
			opt:         WithAuthorizeEventListener(listener, &cancel),
			check: func(assert *assert.Assertions, l *Listener) {
				m.Lock()
				defer m.Unlock()

				count = 0
				_ = l.dispatch(event.Authorize{})
				assert.Equal(1, count)
			},
		}, {
			description: "assert WithAuthorizeEventListener() with cancel work can be canceled",
			r:           validWHR,
			opt:         WithAuthorizeEventListener(listener, &cancel2),
			check: func(assert *assert.Assertions, l *Listener) {
				m.Lock()
				defer m.Unlock()

				cancel2()

				count = 0
				_ = l.dispatch(event.Authorize{})
				assert.Equal(0, count)
			},
		}, {
			description: "assert WithAuthorizeEventListener() without cancel works",
			r:           validWHR,
			opt:         WithAuthorizeEventListener(listener),
			check: func(assert *assert.Assertions, l *Listener) {
				m.Lock()
				defer m.Unlock()

				count = 0
				_ = l.dispatch(event.Authorize{})
				assert.Equal(1, count)
			},
		},
	}
	commonNewTest(t, tests)

	assert.NotNil(t, cancel)
}

func TestWithTokenizeEventListener(t *testing.T) {
	m := sync.Mutex{}
	var cancel CancelEventListenerFunc
	var cancel2 CancelEventListenerFunc

	var count int
	listener := event.TokenizeFunc(func(event.Tokenize) {
		count++
	})

	assert.Nil(t, cancel)

	tests := []newTest{
		{
			description: "assert WithTokenizeEventListener() with cancel works",
			r:           validWHR,
			opt:         WithTokenizeEventListener(listener, &cancel),
			check: func(assert *assert.Assertions, l *Listener) {
				m.Lock()
				defer m.Unlock()

				count = 0
				_ = l.dispatch(event.Tokenize{})
				assert.Equal(1, count)
			},
		}, {
			description: "assert WithTokenizeEventListener() with cancel work can be canceled",
			r:           validWHR,
			opt:         WithTokenizeEventListener(listener, &cancel2),
			check: func(assert *assert.Assertions, l *Listener) {
				m.Lock()
				defer m.Unlock()

				cancel2()

				count = 0
				_ = l.dispatch(event.Tokenize{})
				assert.Equal(0, count)
			},
		}, {
			description: "assert WithTokenizeEventListener() without cancel works",
			r:           validWHR,
			opt:         WithTokenizeEventListener(listener),
			check: func(assert *assert.Assertions, l *Listener) {
				m.Lock()
				defer m.Unlock()

				count = 0
				_ = l.dispatch(event.Tokenize{})
				assert.Equal(1, count)
			},
		},
	}
	commonNewTest(t, tests)

	assert.NotNil(t, cancel)
}

// duplicate of TestWithAuthorizeEventListener(t *testing.T) for RegistrationEventListenerFunc
func TestWithRegistrationEventListener(t *testing.T) {
	m := sync.Mutex{}
	var cancel CancelEventListenerFunc
	var cancel2 CancelEventListenerFunc

	var count int
	listener := event.RegistrationFunc(func(event.Registration) {
		count++
	})

	assert.Nil(t, cancel)

	tests := []newTest{
		{
			description: "assert WithRegistrationEventListener() with cancel works",
			r:           validWHR,
			opt:         WithRegistrationEventListener(listener, &cancel),
			check: func(assert *assert.Assertions, l *Listener) {
				m.Lock()
				defer m.Unlock()

				count = 0
				_ = l.dispatch(event.Registration{})
				assert.Equal(1, count)
			},
		}, {
			description: "assert WithRegistrationEventListener() with cancel work can be canceled",
			r:           validWHR,
			opt:         WithRegistrationEventListener(listener, &cancel2),
			check: func(assert *assert.Assertions, l *Listener) {
				m.Lock()
				defer m.Unlock()

				cancel2()

				count = 0
				_ = l.dispatch(event.Registration{})
				assert.Equal(0, count)
			},
		}, {
			description: "assert WithRegistrationEventListener() without cancel works",
			r:           validWHR,
			opt:         WithRegistrationEventListener(listener),
			check: func(assert *assert.Assertions, l *Listener) {
				m.Lock()
				defer m.Unlock()

				count = 0
				_ = l.dispatch(event.Registration{})
				assert.Equal(1, count)
			},
		},
	}
	commonNewTest(t, tests)

	assert.NotNil(t, cancel)
}
