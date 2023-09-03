// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/xmidt-org/webhook-schema"
	"github.com/xmidt-org/wrp-listener/event"
	"go.uber.org/multierr"
)

const (
	webpaHeader = "X-Webpa-Signature"
	xmidtHeader = "Xmidt-Signature"
)

// Listener provides a way to register a webhook and validate the callbacks.
// It can be configured to register the webhook at a given interval, as well as
// it can also be configured to accept multiple secrets and hash algorithms.
type Listener struct {
	m                     sync.RWMutex
	wg                    sync.WaitGroup
	registration          *webhook.Registration
	webhookURL            string
	registrationOpts      []webhook.Option
	interval              time.Duration
	client                *http.Client
	ctx                   context.Context
	upstreamCtx           context.Context
	shutdown              context.CancelFunc
	update                chan struct{}
	reqDecorators         []Decorator
	registrationListeners listeners
	authorizeListeners    listeners
	tokenizeListeners     listeners
	opts                  []Option
	body                  []byte
	acceptedSecrets       []string
	hashPreferences       []string
	hashes                map[string]func() hash.Hash
}

// Option is an interface that is used to configure the webhook listener.
type Option interface {
	fmt.Stringer
	apply(*Listener) error
}

// New creates a new webhook listener with the given registration and options.
func New(url string, r *webhook.Registration, opts ...Option) (*Listener, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: registration is required", ErrInput)
	}

	url = strings.TrimSpace(url)
	if url == "" {
		return nil, fmt.Errorf("%w: webhook url is required", ErrInput)
	}

	l := Listener{
		registration:     r,
		webhookURL:       url,
		registrationOpts: make([]webhook.Option, 0),
		client:           http.DefaultClient,
		reqDecorators:    make([]Decorator, 0),
		upstreamCtx:      context.Background(),
		ctx:              context.Background(),
		update:           make(chan struct{}, 1),
		acceptedSecrets:  make([]string, 0),
		hashPreferences:  make([]string, 0),
		hashes:           make(map[string]func() hash.Hash, 0),
		opts:             opts,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		err := opt.apply(&l)
		if err != nil {
			return nil, err
		}
	}

	vOpts := []webhook.Option{
		webhook.ValidateRegistrationDuration(0),
	}

	if l.interval != 0 {
		vOpts = append(vOpts, webhook.NoUntil())
	}
	vOpts = append(vOpts, l.registrationOpts...)

	err := l.registration.Validate(vOpts...)
	if err != nil {
		return nil, multierr.Combine(err, fmt.Errorf("%w: invalid registration", ErrInput))
	}

	err = l.use(l.registration.Config.Secret)
	if err != nil {
		return nil, err
	}

	return &l, nil
}

// AddRegistrationEventListener adds an event listener to the webhook listener.
// The listener will be called for each event that occurs.  The returned
// function can be called to remove the listener.
func (l *Listener) AddRegistrationEventListener(listener event.RegistrationListener) CancelEventListenerFunc {
	return l.registrationListeners.addListener(listener)
}

// AddTokenizeEventListener adds an event listener to the webhook listener.
// The listener will be called for each event that occurs.  The returned
// function can be called to remove the listener.
func (l *Listener) AddTokenizeEventListener(listener event.TokenizeListener) CancelEventListenerFunc {
	return l.tokenizeListeners.addListener(listener)
}

// AddAuthorizeEventListener adds an event listener to the webhook listener.
// The listener will be called for each event that occurs.  The returned
// function can be called to remove the listener.
func (l *Listener) AddAuthorizeEventListener(listener event.AuthorizeListener) CancelEventListenerFunc {
	return l.authorizeListeners.addListener(listener)
}

// dispatch dispatches the event to the listeners and returns the error that
// should be returned by the caller.
func (l *Listener) dispatch(evnt any) error {
	switch evnt := evnt.(type) {
	case event.Registration:
		l.registrationListeners.visit(func(listener any) {
			listener.(event.RegistrationListener).OnRegistrationEvent(evnt)
		})
		return evnt.Err
	case event.Tokenize:
		l.tokenizeListeners.visit(func(listener any) {
			listener.(event.TokenizeListener).OnTokenizeEvent(evnt)
		})
		return evnt.Err
	case event.Authorize:
		l.authorizeListeners.visit(func(listener any) {
			listener.(event.AuthorizeListener).OnAuthorizeEvent(evnt)
		})
		return evnt.Err
	}

	panic("unknown event type")
}

// Register registers the webhook listener using the optional specified secret.
// If the interval is greater than 0 the registrations will continue until
// Stop() is called or the parent context is canceled.  If the listener is
// already running, the secret will be updated immediately.
// If the secret is not provided, the current secret will be used.  Only the
// first secret will be used if multiple secrets are provided.
func (l *Listener) Register(secret ...string) error {
	l.m.Lock()
	defer l.m.Unlock()

	if len(secret) != 0 {
		if err := l.use(secret[0]); err != nil {
			return err
		}
	}

	if l.shutdown != nil {
		return nil
	}

	if l.interval == 0 {
		return l.register(true)
	}

	l.ctx, l.shutdown = context.WithCancel(l.upstreamCtx)
	go l.run()

	return nil
}

// Stop stops the webhook listener.  If the listener is not running, this is a
// no-op.
func (l *Listener) Stop() {
	l.m.Lock()
	shutdown := l.shutdown
	l.m.Unlock()

	if shutdown != nil {
		shutdown()
	}
	l.wg.Wait()
}

func (l *Listener) use(secret string) error {
	l.registration.Config.Secret = secret

	var err error
	l.body, err = json.Marshal(&l.registration)
	if err != nil {
		return multierr.Combine(err, fmt.Errorf("%w: unable to marshal the registration", ErrInput))
	}

	// Update the hash functions without blocking.
	select {
	case l.update <- struct{}{}:
	default:
	}

	return nil
}

// Accept defines the entire list of secrets to accept for the webhook callbacks.
// If any of the secrets match the secret in the token, the request will be
// authorized.
func (l *Listener) Accept(secrets []string) {
	l.m.Lock()
	defer l.m.Unlock()

	l.acceptedSecrets = make([]string, len(secrets))
	copy(l.acceptedSecrets, secrets)
}

// run is the main loop for the webhook listener.  It will register the webhook
// at the given interval until Stop() is called.
func (l *Listener) run() {
	l.wg.Add(1)
	defer l.wg.Done()

	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	for {
		_ = l.register(false)

		select {
		case <-l.ctx.Done():
			return

		case <-ticker.C:
		case <-l.update:
		}
	}
}

// String returns a string representation of the webhook listener and the options
// used to configure it.
func (l *Listener) String() string {
	buf := strings.Builder{}

	buf.WriteString("Listener(")
	buf.WriteString("URL(")
	buf.WriteString(l.webhookURL)
	buf.WriteString(")")

	for _, opt := range l.opts {
		if opt == nil {
			continue
		}
		buf.WriteString(", ")
		buf.WriteString(opt.String())
	}
	buf.WriteString(")")

	return buf.String()
}

// register registers the webhook listener.  The newest secret will be used for
// the registration.  The locked argument determines if a mutex is already held
// by the caller to prevent deadlock.
func (l *Listener) register(locked bool) error {
	// Keep the lock block as small as possible.  Copy out the values that are
	// needed and release the lock.
	if !locked {
		l.m.RLock()
	}

	address := l.webhookURL
	body := l.body

	if !locked {
		l.m.RUnlock()
	}

	var evnt event.Registration

	req, err := http.NewRequest(http.MethodPost, address, bytes.NewReader(body))
	if err != nil {
		evnt.Err = multierr.Combine(err, ErrNewRequestFailed, ErrRegistrationNotAttempted)
		return l.dispatch(evnt)
	}

	for _, decorator := range l.reqDecorators {
		err := decorator.Decorate(req)
		if err != nil {
			evnt.Err = multierr.Combine(err, ErrDecoratorFailed, ErrRegistrationNotAttempted)
			return l.dispatch(evnt)
		}
	}

	req.Header.Set("Content-Type", "application/json")

	evnt.At = time.Now()
	resp, err := l.client.Do(req)
	evnt.Duration = time.Since(evnt.At)

	if err != nil {
		evnt.Err = multierr.Combine(err, ErrRegistrationFailed)
		return l.dispatch(evnt)
	}
	defer resp.Body.Close()

	evnt.StatusCode = resp.StatusCode

	if resp.StatusCode == http.StatusOK {
		evnt.Until = evnt.At.Add(time.Duration(l.registration.Duration))
		return l.dispatch(evnt)
	}

	evnt.Body, _ = io.ReadAll(resp.Body)
	evnt.Err = ErrRegistrationFailed

	return l.dispatch(evnt)
}

// Tokenize parses the token from the request header.  If the token is not found
// or is invalid, an error is returned.
func (l *Listener) Tokenize(r *http.Request) (*token, error) {
	evnt := event.Tokenize{
		Header: xmidtHeader,
	}

	headers := r.Header.Values(xmidtHeader)
	if len(headers) == 0 {
		headers = r.Header.Values(webpaHeader)
		evnt.Header = webpaHeader
	}

	if len(headers) == 0 {
		evnt.Header = ""
	}

	choices := map[string]string{
		"none": "",
	}
	list := make([]string, 0, len(headers))
	list = append(list, "none")

	for _, header := range headers {
		header = strings.TrimSpace(header)
		if header == "" {
			continue
		}
		parts := strings.Split(header, "=")
		if len(parts) != 2 {
			evnt.Err = multierr.Combine(ErrInvalidTokenHeader, ErrInvalidHeaderFormat)
			return nil, l.dispatch(evnt)
		}

		alg := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		if alg == "" || val == "" {
			evnt.Err = multierr.Combine(ErrInvalidTokenHeader, ErrInvalidHeaderFormat)
			return nil, l.dispatch(evnt)
		}

		choices[alg] = val
		list = append(list, alg)
	}

	evnt.Algorithms = list
	best, err := l.best(list)
	if err != nil {
		evnt.Err = multierr.Combine(ErrInvalidTokenHeader, ErrAlgorithmNotFound)
		return nil, l.dispatch(evnt)
	}

	evnt.Algorithm = best
	l.dispatch(evnt)
	return newToken(best, choices[best]), nil
}

// Authorize validates that the request body matches the hash and secret provided
// in the token.
func (l *Listener) Authorize(r *http.Request, t Token) error {
	var evnt event.Authorize

	if t == nil {
		evnt.Err = ErrNoToken
		return l.dispatch(evnt)
	}

	secret, err := hex.DecodeString(t.Principal())
	if err != nil {
		evnt.Err = multierr.Combine(err, ErrInvalidSignature)
		return l.dispatch(evnt)
	}

	var msg []byte
	if r.Body != nil {
		msg, err = io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			evnt.Err = multierr.Combine(err, ErrUnableToReadBody)
			return l.dispatch(evnt)
		}

		// Reset the body so it can be read again later.
		r.Body = io.NopCloser(bytes.NewReader(msg))
	}

	evnt.Algorithm = t.Type()
	hashes, err := l.getHashes(evnt.Algorithm)
	if err != nil {
		evnt.Err = err
		return l.dispatch(evnt)
	}

	for _, h := range hashes {
		h.Write(msg)
		if hmac.Equal(h.Sum(nil), secret) {
			l.dispatch(evnt)
			return nil
		}
	}

	evnt.Err = ErrInvalidSignature
	return l.dispatch(evnt)
}

// best returns the best secret to use for the given choices.  If none of the
// choices are in the list of secrets, an empty string is returned.
func (l *Listener) best(choices []string) (string, error) {
	l.m.RLock()
	defer l.m.RUnlock()

	for _, want := range l.hashPreferences {
		for _, choice := range choices {
			if choice == want {
				return choice, nil
			}
		}
	}

	return "", ErrNotAcceptedHash
}

// hashes returns a slice of hashes of the active secrets.
func (l *Listener) getHashes(which string) ([]hash.Hash, error) {
	l.m.RLock()
	defer l.m.RUnlock()

	h, found := l.hashes[which]
	if !found {
		return nil, ErrNotAcceptedHash
	}

	hashes := make([]hash.Hash, 0, len(l.acceptedSecrets))
	for _, secret := range l.acceptedSecrets {
		hashes = append(hashes, hmac.New(h, []byte(secret)))
	}
	return hashes, nil
}
