// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"context"
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/webhook-schema"
	"go.uber.org/zap"
)

func TestOptionStrings(t *testing.T) {
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
			in:       Logger(zap.NewNop()),
			expected: "Logger(zap)",
		}, {
			in:       Logger(nil),
			expected: "Logger(nil)",
		}, {
			in:       Metrics(new(Measure)),
			expected: "Metrics(metrics)",
		}, {
			in:       Metrics(nil),
			expected: "Metrics(nil)",
		}, {
			in:       HTTPClient(http.DefaultClient),
			expected: "HTTPClient(client)",
		}, {
			in:       HTTPClient(nil),
			expected: "HTTPClient(nil)",
		}, {
			in:       AuthBasic("user", "pass"),
			expected: "AuthBasic(user, ***)",
		}, {
			in:       AuthBasicFunc(func() (string, string, error) { return "", "", nil }),
			expected: "AuthBasicFunc(fn)",
		}, {
			in:       AuthBasicFunc(nil),
			expected: "AuthBasicFunc(nil)",
		}, {
			in:       AuthBearer("secret_token"),
			expected: "AuthBearer(***)",
		}, {
			in:       AuthBearerFunc(func() (string, error) { return "", nil }),
			expected: "AuthBearerFunc(fn)",
		}, {
			in:       AuthBearerFunc(nil),
			expected: "AuthBearerFunc(nil)",
		}, {
			in:       Secret("foo"),
			expected: "Secret(***)",
		}, {
			in:       Secrets("foo", "bar"),
			expected: "Secrets(***, ...)",
		}, {
			in:       AcceptNone(),
			expected: "AcceptNone()",
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
		},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.in.String())
		})
	}
}

func TestLogger(t *testing.T) {
	logger, err := zap.NewProduction()
	require.NoError(t, err)

	tests := []newTest{
		{
			description: "assert default logger is there",
			r:           validWHR,
			check: func(assert *assert.Assertions, l *Listener) {
				assert.NotNil(l.logger)
			},
		}, {
			description: "assert specified logger is there",
			r:           validWHR,
			opt:         Logger(logger),
			check: func(assert *assert.Assertions, l *Listener) {
				assert.Equal(l.logger, logger)
			},
		}, {
			description: "assert nil logger is there",
			r:           validWHR,
			opts:        []Option{Logger(logger), Logger(nil)},
			check: func(assert *assert.Assertions, l *Listener) {
				assert.NotEqual(l.logger, logger)
			},
		}, {
			description: "nearly empty with an invalid interval",
			opt:         Interval(-5 * time.Minute),
			expectedErr: ErrInput,
		},
	}
	commonNewTest(t, tests)
}

func TestMetrics(t *testing.T) {
	metrics := Measure{
		Registration: MeasureRegistration{
			Label: "label",
		},
	}

	tests := []newTest{
		{
			description: "assert default metrics is there",
			r:           validWHR,
			check: func(assert *assert.Assertions, l *Listener) {
				tmp := new(Measure).init()
				assert.Equal(l.metrics, tmp)
			},
		}, {
			description: "assert new metrics are used",
			r:           validWHR,
			opt:         Metrics(&metrics),
			check: func(assert *assert.Assertions, l *Listener) {
				tmp := metrics
				tmp.init()
				assert.Equal(l.metrics, &tmp)
			},
		}, {
			description: "assert nil metrics works",
			r:           validWHR,
			opts:        []Option{Metrics(&metrics), Metrics(nil)},
			check: func(assert *assert.Assertions, l *Listener) {
				tmp := new(Measure).init()
				assert.Equal(l.metrics, tmp)
			},
		},
	}
	commonNewTest(t, tests)
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

func TestAuth(t *testing.T) {
	tests := []newTest{
		{
			description: "assert default auth is empty",
			r:           validWHR,
			check:       vadorGetAuth(""),
		}, {
			description: "assert AuthBasic works",
			r:           validWHR,
			opt:         AuthBasic("user", "pass"),
			check:       vadorGetAuth("Basic dXNlcjpwYXNz"),
		}, {
			description: "assert AuthBasicFunc works",
			r:           validWHR,
			opt: AuthBasicFunc(func() (string, string, error) {
				return "user", "pass", nil
			}),
			check: vadorGetAuth("Basic dXNlcjpwYXNz"),
		}, {
			description: "assert AuthBasicFunc handles failure",
			r:           validWHR,
			opt: AuthBasicFunc(func() (string, string, error) {
				return "", "", ErrInput
			}),
			check: vadorGetAuth("", ErrInput),
		}, {
			description: "assert AuthBasicFunc(nil) works",
			r:           validWHR,
			opt:         AuthBasicFunc(nil),
			check:       vadorGetAuth(""),
		}, {
			description: "assert AuthBearer works",
			r:           validWHR,
			opt:         AuthBearer("token"),
			check:       vadorGetAuth("Bearer token"),
		}, {
			description: "assert AuthBearerFunc works",
			r:           validWHR,
			opt: AuthBearerFunc(func() (string, error) {
				return "token", nil
			}),
			check: vadorGetAuth("Bearer token"),
		}, {
			description: "assert AuthBearerFunc(nil) works",
			r:           validWHR,
			opt:         AuthBearerFunc(nil),
			check:       vadorGetAuth(""),
		}, {
			description: "assert AuthBearerFunc handles failure",
			r:           validWHR,
			opt: AuthBearerFunc(func() (string, error) {
				return "", ErrInput
			}),
			check: vadorGetAuth("", ErrInput),
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
			description: "assert Secret() works",
			r:           validWHR,
			opt:         Secret("foo"),
			check:       vadorAcceptedSecrets("foo"),
		}, {
			description: "assert Secrets() works",
			r:           validWHR,
			opt:         Secrets("foo", "bar"),
			check:       vadorAcceptedSecrets("foo", "bar"),
		}, {
			description: "assert multiple Secret(), Secrets() works",
			r:           validWHR,
			opts: []Option{
				Secret("foo"),
				Secrets("car", "cat"),
				Secret("bar"),
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
			opt:         AcceptNone(),
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
