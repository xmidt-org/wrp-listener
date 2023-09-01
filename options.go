// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"context"
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"fmt"
	"hash"
	"net/http"
	"strings"
	"time"

	"github.com/xmidt-org/webhook-schema"
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

// The Decorator type is an adapter to allow the use of ordinary functions as
// decorators.  If f is a function with the appropriate signature,
// DecoratorFunc(f) is a Decorator that calls f.
type DecoratorFunc func(*http.Request) error

func (d DecoratorFunc) Decorate(r *http.Request) error {
	return d(r)
}

func (d DecoratorFunc) String() string {
	return "DecoratorFunc(fn)"
}

// A Decorator decorates an http request before it is sent to the webhook
// registration endpoint.
type Decorator interface {
	fmt.Stringer
	Decorate(*http.Request) error
}

// DecorateRequest is an option that provides the function to use to decorate
// the http request before it is sent to the webhook registration endpoint.
// This is useful for adding headers or other information to the request.
//
// Examples of this include adding an authorization header, or additional
// headers to the request.
//
// Multiple DecorateRequest options can be provided.  They will be called in
// the order they are provided.
func DecorateRequest(d Decorator) Option {
	return &decorateRequestOption{
		d: d,
	}
}

type decorateRequestOption struct {
	d Decorator
}

func (d decorateRequestOption) apply(lis *Listener) error {
	if d.d != nil {
		lis.reqDecorators = append(lis.reqDecorators, d.d)
	}
	return nil
}

func (d decorateRequestOption) String() string {
	if d.d != nil {
		return "DecorateRequest(" + d.d.String() + ")"
	}
	return "DecorateRequest(nil)"
}

// AcceptedSecrets is an option that provides the list of secrets accepted
// by the webhook listener when validating the callback event.  A valid
// hash (or multiple) must be provided as well.
func AcceptedSecrets(secrets ...string) Option {
	return &acceptedSecretsOption{
		secret: secrets,
	}
}

type acceptedSecretsOption struct {
	secret []string
}

func (s acceptedSecretsOption) apply(lis *Listener) error {
	lis.acceptedSecrets = append(lis.acceptedSecrets, s.secret...)
	return nil
}

func (s acceptedSecretsOption) String() string {
	if len(s.secret) == 1 {
		return "AcceptedSecrets(***)"
	}
	return "AcceptedSecrets(***, ...)"
}

// AcceptNoHash enables the use of no hash for the webhook listener
// callback validation.
//
// USE WITH CAUTION.
func AcceptNoHash() Option {
	return &hashOption{
		text: "AcceptNoHash()",
		name: "none",
	}
}

// AcceptSHA1 enables the use of the sha1 hash for the webhook listener
// callback validation.
func AcceptSHA1() Option {
	return &hashOption{
		text: "AcceptSHA1()",
		name: "sha1",
		fn:   sha1.New,
	}
}

// AcceptSHA256 enables the use of the sha256 hash for the webhook listener
// callback validation.
func AcceptSHA256() Option {
	return &hashOption{
		text: "AcceptSHA256()",
		name: "sha256",
		fn:   sha256.New,
	}
}

// AcceptCustom is an option that sets the hash to use for the webhook
// callback validation to use.  A nil hash is not accepted.
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

// WebhookOpts is an option that provides the webhook.Options to apply
// during the validation of the registration of the webhook.
func WebhookOpts(opts ...webhook.Option) Option {
	return &webhookOptsOption{
		opts: opts,
	}
}

type webhookOptsOption struct {
	opts []webhook.Option
}

func (r webhookOptsOption) apply(lis *Listener) error {
	lis.registrationOpts = append(lis.registrationOpts, r.opts...)
	return nil
}

func (r webhookOptsOption) String() string {
	buf := strings.Builder{}

	buf.WriteString("RegistrationOpts(")
	for i, opt := range r.opts {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(opt.String())
	}
	buf.WriteString(")")
	return buf.String()
}

// Context is an option that provides the context to use for the webhook
// listener registration to use.  A nil value will cause the default context
// to be used.
func Context(ctx context.Context) Option {
	return &contextOption{
		ctx: ctx,
	}
}

type contextOption struct {
	ctx context.Context
}

func (c contextOption) apply(lis *Listener) error {
	lis.upstreamCtx = c.ctx
	return nil
}

func (c contextOption) String() string {
	return "Context(ctx)"
}
