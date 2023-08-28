// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Interval is an option that sets the interval to wait between webhook
// registration attempts.  The default is to only register once.  This option
// must be greater than or equal to 0.  A value of 0 will cause the webhook to
// only be registered once.
func Interval(i time.Duration) Option {
	return &intervalOption{
		text:     fmt.Sprintf("Interval(%s)", i),
		interval: i,
	}
}

// Once is an option that sets the webhook to only be registered once.  This is
// the default behavior.
func Once() Option {
	return &intervalOption{
		text: "Once()",
	}
}

type intervalOption struct {
	text     string
	interval time.Duration
}

func (i intervalOption) apply(l *Listener) error {
	if i.interval < 0 {
		return fmt.Errorf("%w, interval must be greater than 0", ErrInput)
	}

	l.interval = i.interval
	return nil
}

func (i intervalOption) String() string {
	return i.text
}

// Logger is an option that sets the logger to use for the webhook listener.
// The default is to use a no-op logger.
func Logger(l *zap.Logger) Option {
	return &loggerOption{
		logger: l,
	}
}

type loggerOption struct {
	logger *zap.Logger
}

func (l loggerOption) apply(lis *Listener) error {
	logger := l.logger
	if logger == nil {
		logger = zap.NewNop()
	}

	lis.logger = logger
	return nil
}

func (l loggerOption) String() string {
	if l.logger != nil {
		return "Logger(zap)"
	}
	return "Logger(nil)"
}

// Metrics is an option that provides the metrics to use for the webhook listener.
// The default is to not use metrics.  Any metrics that are not provided will be
// replaced with a no-op metric.  If a nil value is provided, then a no-op
// metric will be used.
func Metrics(m *Measure) Option {
	m.init()
	return &metricsOption{
		metrics: m,
	}
}

type metricsOption struct {
	metrics *Measure
}

func (m metricsOption) apply(lis *Listener) error {
	metrics := m.metrics
	if m.metrics == nil {
		metrics = new(Measure).init()
	}
	lis.metrics = metrics
	return nil
}

func (m metricsOption) String() string {
	if m.metrics != nil {
		return "Metrics(metrics)"
	}
	return "Metrics(nil)"
}

// HTTPClient is an option that provides the http client to use for the
// webhook listener registration to use.  A nil value will cause the default
// http client to be used.
func HTTPClient(c *http.Client) Option {
	return &httpClientOption{
		client: c,
	}
}

type httpClientOption struct {
	client *http.Client
}

func (h httpClientOption) apply(lis *Listener) error {
	if h.client == nil {
		lis.client = http.DefaultClient
		return nil
	}

	lis.client = h.client
	return nil
}

func (h httpClientOption) String() string {
	if h.client != nil {
		return "HTTPClient(client)"
	}
	return "HTTPClient(nil)"
}

// AuthBasic is an option that provides the basic auth credentials to use
// for the webhook listener registration to use.
func AuthBasic(username, password string) Option {
	return &authFuncOption{
		text: fmt.Sprintf("AuthBasic(%s, ***)", username),
		fn: func() (string, error) {
			return basicToCredentials(username, password), nil
		},
	}
}

// AuthBasicFunc is an option that provides a function that will be called
// to get the basic auth credentials to use for the webhook listener
// registration to use.  A nil value will cause no credentials to be used.
func AuthBasicFunc(fn func() (username string, password string, err error)) Option {
	if fn == nil {
		return &authFuncOption{
			text: "AuthBasicFunc(nil)",
			fn: func() (string, error) {
				return "", nil
			},
		}
	}

	return &authFuncOption{
		text: "AuthBasicFunc(fn)",
		fn: func() (string, error) {
			username, password, err := fn()
			if err != nil {
				return "", err
			}
			return basicToCredentials(username, password), nil
		},
	}
}

// AuthBearerFunc is an option that provides a function that will be called
// to get the bearer auth token to use for the webhook listener registration to
// use.  A nil value will cause no credentials to be used.
func AuthBearerFunc(fn func() (string, error)) Option {
	if fn == nil {
		return &authFuncOption{
			text: "AuthBearerFunc(nil)",
			fn: func() (string, error) {
				return "", nil
			},
		}
	}

	return &authFuncOption{
		text: "AuthBearerFunc(fn)",
		fn: func() (string, error) {
			token, err := fn()
			if err != nil {
				return "", err
			}
			return "Bearer " + token, nil
		},
	}
}

// AuthBearer is an option that provides the bearer auth token to use for
// the webhook listener registration to use.  An empty value will cause no
// credentials to be used.
func AuthBearer(token string) Option {
	return &authFuncOption{
		text: "AuthBearer(***)",
		fn: func() (string, error) {
			return "Bearer " + token, nil
		},
	}
}

func basicToCredentials(username, password string) string {
	credentials := username + ":" + password
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	return "Basic " + encodedCredentials
}

type authFuncOption struct {
	text string
	fn   func() (string, error)
}

func (a authFuncOption) apply(lis *Listener) error {
	lis.getAuth = a.fn
	return nil
}

func (a authFuncOption) String() string {
	return a.text
}

// Secret is an option that provides the secret to use for the webhook
// listener registration to use.  A nil value will cause no secret to be used.
func Secret(secret string) Option {
	return &secretFuncOption{
		text:   "Secret(***)",
		secret: []string{secret},
	}
}

func Secrets(secrets ...string) Option {
	return &secretFuncOption{
		text:   "Secrets(***, ...)",
		secret: secrets,
	}
}

type secretFuncOption struct {
	text   string
	secret []string
}

func (s secretFuncOption) apply(lis *Listener) error {
	lis.acceptedSecrets = append(lis.acceptedSecrets, s.secret...)
	return nil
}

func (s secretFuncOption) String() string {
	return s.text
}

// AcceptNone enables the use of the sha1 hash for the webhook listener
// validation to use.
func AcceptNone() Option {
	return &hashOption{
		text: "AcceptNone()",
		name: "none",
	}
}

// AcceptSHA1 enables the use of the sha1 hash for the webhook listener
// validation to use.
func AcceptSHA1() Option {
	return &hashOption{
		text: "AcceptSHA1()",
		name: "sha1",
		fn:   sha1.New,
	}
}

// AcceptSHA256 enables the use of the sha256 hash for the webhook listener
// validation to use.
func AcceptSHA256() Option {
	return &hashOption{
		text: "AcceptSHA256()",
		name: "sha256",
		fn:   sha256.New,
	}
}

// AcceptCustom is an option that sets the hash to use for the webhook
// listener validation to use.  A nil value will cause no hash to be used.
func AcceptCustom(name string, h func() hash.Hash) Option {
	if h == nil {
		return &hashOption{
			err: fmt.Errorf("%w, hash function cannot be nil", ErrInput),
		}
	}

	return &hashOption{
		text: "AcceptCustom(" + name + ", fn)",
		name: name,
		fn:   h,
	}
}

type hashOption struct {
	name string
	text string
	fn   func() hash.Hash
	err  error
}

func (h hashOption) apply(lis *Listener) error {
	if h.err != nil {
		return h.err
	}

	lis.hashPreferences = append(lis.hashPreferences, h.name)
	lis.hashes[h.name] = h.fn
	return nil
}

func (h hashOption) String() string {
	return h.text
}
