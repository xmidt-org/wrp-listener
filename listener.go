// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/xmidt-org/eventor"
	"github.com/xmidt-org/webhook-schema"
	"github.com/xmidt-org/wrp-listener/event"
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
	shutdown              context.CancelFunc
	update                chan struct{}
	reqDecorators         []Decorator
	registrationListeners eventor.Eventor[event.RegistrationListener]
	authorizeListeners    eventor.Eventor[event.AuthorizeListener]
	tokenizeListeners     eventor.Eventor[event.TokenizeListener]
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

// CancelEventListenerFunc removes the listener it's associated with and cancels any
// future events sent to that listener.
//
// A CancelEventListenerFunc is idempotent:  after the first invocation, calling this
// closure will have no effect.
type CancelEventListenerFunc func()

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
		return nil, errors.Join(err, fmt.Errorf("%w: invalid registration", ErrInput))
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
	return CancelEventListenerFunc(l.registrationListeners.Add(listener))
}

// AddTokenizeEventListener adds an event listener to the webhook listener.
// The listener will be called for each event that occurs.  The returned
// function can be called to remove the listener.
func (l *Listener) AddTokenizeEventListener(listener event.TokenizeListener) CancelEventListenerFunc {
	return CancelEventListenerFunc(l.tokenizeListeners.Add(listener))
}

// AddAuthorizeEventListener adds an event listener to the webhook listener.
// The listener will be called for each event that occurs.  The returned
// function can be called to remove the listener.
func (l *Listener) AddAuthorizeEventListener(listener event.AuthorizeListener) CancelEventListenerFunc {
	return CancelEventListenerFunc(l.authorizeListeners.Add(listener))
}

// dispatch dispatches the event to the listeners and returns the error that
// should be returned by the caller.
func dispatch[T event.Authorize | event.Registration | event.Tokenize](l *Listener, evnt T) error {
	var err error
	switch evnt := any(evnt).(type) {
	case event.Registration:
		l.registrationListeners.Visit(func(listener event.RegistrationListener) {
			listener.OnRegistrationEvent(evnt)
		})
		err = evnt.Err
	case event.Tokenize:
		l.tokenizeListeners.Visit(func(listener event.TokenizeListener) {
			listener.OnTokenizeEvent(evnt)
		})
		err = evnt.Err
	case event.Authorize:
		l.authorizeListeners.Visit(func(listener event.AuthorizeListener) {
			listener.OnAuthorizeEvent(evnt)
		})
		err = evnt.Err
	}
	return err
}

// Register registers the webhook listener using the optional specified secret.
// If the interval is greater than 0 the registrations will continue until
// Stop() is called or the parent context is canceled.  If the listener is
// already running, the secret will be updated immediately.
// If the secret is not provided, the current secret will be used.  Only the
// first secret will be used if multiple secrets are provided.
func (l *Listener) Register(ctx context.Context, secret ...string) error {
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
		_, err := l.register(ctx, true, time.Time{})
		return err
	}

	ctx, l.shutdown = context.WithCancel(ctx)
	go l.run(ctx)

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
		return errors.Join(err, fmt.Errorf("%w: unable to marshal the registration", ErrInput))
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
func (l *Listener) run(ctx context.Context) {
	var presentExpiration time.Time
	l.wg.Add(1)
	defer l.wg.Done()

	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	for {
		exp, err := l.register(ctx, false, presentExpiration)
		if err == nil {
			presentExpiration = exp
			ticker.Reset(l.interval)
		} else {
			// TODO add better retry logic
			ticker.Reset(time.Second)
		}

		select {
		case <-ctx.Done():
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
func (l *Listener) register(ctx context.Context, locked bool, presentExpiration time.Time) (time.Time, error) {
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

	evnt := event.Registration{
		Until: presentExpiration,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, address, bytes.NewReader(body))
	if err != nil {
		evnt.Err = errors.Join(err, ErrNewRequestFailed, ErrRegistrationNotAttempted)
		return time.Time{}, dispatch(l, evnt)
	}

	for _, decorator := range l.reqDecorators {
		err := decorator.Decorate(req)
		if err != nil {
			evnt.Err = errors.Join(err, ErrDecoratorFailed, ErrRegistrationNotAttempted)
			return time.Time{}, dispatch(l, evnt)
		}
	}

	req.Header.Set("Content-Type", "application/json")

	evnt.At = time.Now()
	resp, err := l.client.Do(req)
	evnt.Duration = time.Since(evnt.At)

	if err != nil {
		evnt.Err = errors.Join(err, ErrRegistrationFailed)
		return time.Time{}, dispatch(l, evnt)
	}
	defer resp.Body.Close()

	evnt.StatusCode = resp.StatusCode

	if resp.StatusCode == http.StatusOK {
		evnt.Until = evnt.At.Add(time.Duration(l.registration.Duration))
		return evnt.Until, dispatch(l, evnt)
	}

	evnt.Body, _ = io.ReadAll(resp.Body)
	evnt.Err = ErrRegistrationFailed

	return time.Time{}, dispatch(l, evnt)
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
			evnt.Err = errors.Join(ErrInvalidTokenHeader, ErrInvalidHeaderFormat)
			return nil, dispatch(l, evnt)
		}

		alg := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		if alg == "" || val == "" {
			evnt.Err = errors.Join(ErrInvalidTokenHeader, ErrInvalidHeaderFormat)
			return nil, dispatch(l, evnt)
		}

		choices[alg] = val
		list = append(list, alg)
	}

	evnt.Algorithms = list
	best, err := l.best(list)
	if err != nil {
		evnt.Err = errors.Join(ErrInvalidTokenHeader, ErrAlgorithmNotFound)
		return nil, dispatch(l, evnt)
	}

	evnt.Algorithm = best
	dispatch(l, evnt)
	return newToken(best, choices[best]), nil
}

// Authorize validates that the request body matches the hash and secret provided
// in the token.
func (l *Listener) Authorize(r *http.Request, t Token) error {
	var evnt event.Authorize

	if t == nil {
		evnt.Err = ErrNoToken
		return dispatch(l, evnt)
	}

	secret, err := hex.DecodeString(t.Principal())
	if err != nil {
		evnt.Err = errors.Join(err, ErrInvalidSignature)
		return dispatch(l, evnt)
	}

	var msg []byte
	if r.Body != nil {
		msg, err = io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			evnt.Err = errors.Join(err, ErrUnableToReadBody)
			return dispatch(l, evnt)
		}

		// Reset the body so it can be read again later.
		r.Body = io.NopCloser(bytes.NewReader(msg))
	}

	evnt.Algorithm = t.Type()
	hashes, err := l.getHashes(evnt.Algorithm)
	if err != nil {
		evnt.Err = err
		return dispatch(l, evnt)
	}

	for _, h := range hashes {
		h.Write(msg)
		if hmac.Equal(h.Sum(nil), secret) {
			dispatch(l, evnt)
			return nil
		}
	}

	evnt.Err = ErrInvalidSignature
	return dispatch(l, evnt)
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
