// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"bytes"
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
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	webpaHeader = "X-Webpa-Signature"
	xmidtHeader = "Xmidt-Signature"
)

// Listener provides a way to register a webhook and validate the callbacks.
// It can be configured to register the webhook at a given interval, as well as
// it can also be configured to accept multiple secrets and hash algorithms.
type Listener struct {
	m                sync.RWMutex
	registration     *webhook.Registration
	registrationOpts []webhook.Option
	interval         time.Duration
	client           *http.Client
	shutdown         chan struct{}
	update           chan struct{}
	running          bool
	wg               sync.WaitGroup
	getAuth          func() (string, error)
	logger           *zap.Logger
	metrics          *Measure
	opts             []Option
	body             []byte
	acceptedSecrets  []string
	hashPreferences  []string
	hashes           map[string]func() hash.Hash
}

// Option is an interface that is used to configure the webhook listener.
type Option interface {
	fmt.Stringer
	apply(*Listener) error
}

// New creates a new webhook listener with the given registration and options.
func New(r *webhook.Registration, opts ...Option) (*Listener, error) {
	l := Listener{
		registration:     r,
		registrationOpts: make([]webhook.Option, 0),
		logger:           zap.NewNop(),
		client:           http.DefaultClient,
		getAuth:          func() (string, error) { return "", nil },
		shutdown:         make(chan struct{}),
		update:           make(chan struct{}, 1),
		metrics:          new(Measure).init(),
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

	l.use(l.registration.Config.Secret)

	return &l, nil
}

// Register registers the webhook listener.  If the interval is greater than 0
// the registrations will continue until Stop() is called.  If the listener is
// already running, this is a nop.
func (l *Listener) Register() error {
	l.m.Lock()
	defer l.m.Unlock()

	if l.running {
		return nil
	}

	if l.interval == 0 {
		return l.register(true)
	}

	l.wg.Add(1)
	go l.run()
	l.running = true

	return nil
}

// Stop stops the webhook listener.  If the listener is not running, this is a
// no-op.
func (l *Listener) Stop() {
	l.m.Lock()

	if !l.running {
		l.m.Unlock()
		return
	}

	close(l.shutdown)
	l.running = false

	l.m.Unlock()

	l.wg.Wait()
}

// Use sets the secret to use for the webhook registration.  If the listener is
// running, the registration will be updated immediately.
func (l *Listener) Use(secret string) error {
	l.m.Lock()
	defer l.m.Unlock()

	return l.use(secret)
}

func (l *Listener) use(secret string) error {
	l.registration.Config.Secret = secret

	var err error
	l.body, err = json.Marshal(&l.registration)
	if err != nil {
		return multierr.Combine(err, fmt.Errorf("%w: unable to marshal the registration", ErrInput))
	}

	if l.running {
		l.update <- struct{}{}
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
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()
	defer l.wg.Done()

	l.metrics.RegistrationInterval.Set(l.interval.Seconds())

	for {
		select {
		case <-l.shutdown:
			return
		case <-ticker.C:
			if err := l.register(false); err != nil {
				l.logger.Error("failed to register webhook", zap.Error(err))
			}
		case <-l.update:
			if err := l.register(false); err != nil {
				l.logger.Error("failed to register webhook", zap.Error(err))
			}
		}
	}
}

// String returns a string representation of the webhook listener and the options
// used to configure it.
func (l *Listener) String() string {
	buf := strings.Builder{}

	buf.WriteString("Listener(")
	comma := ""
	for _, opt := range l.opts {
		if opt == nil {
			continue
		}
		buf.WriteString(comma)
		buf.WriteString(opt.String())
		comma = ", "
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

	fn := l.getAuth
	address := l.registration.Address
	body := l.body

	if !locked {
		l.m.RUnlock()
	}

	auth, err := fn()
	if err != nil {
		l.metrics.Registration.incAuthFetchingFailure()
		return multierr.Combine(err, fmt.Errorf("%w: unable to fetch auth", ErrRegistrationNotAttempted))
	}

	req, err := http.NewRequest(http.MethodPost, address, bytes.NewReader(body))
	if err != nil {
		l.metrics.Registration.incNewRequestFailure()
		return multierr.Combine(err, fmt.Errorf("%w: unable to create a new http request", ErrRegistrationNotAttempted))
	}

	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := l.client.Do(req)
	if err != nil {
		l.metrics.Registration.incRequestFailure()
		return multierr.Combine(err, fmt.Errorf("%w: unable to make a http request", ErrRegistrationFailed))
	}
	defer resp.Body.Close()

	l.metrics.Registration.incStatusCode(resp.StatusCode)

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	rBody, _ := io.ReadAll(resp.Body)

	l.logger.Info("failed to register webhook",
		zap.Int("StatusCode", resp.StatusCode),
		zap.ByteString("body", rBody))

	return fmt.Errorf("%w: http status code(%d) was not 200",
		ErrRegistrationFailed, resp.StatusCode)
}

// Tokenize parses the token from the request header.  If the token is not found
// or is invalid, an error is returned.
func (l *Listener) Tokenize(r *http.Request) (*Token, error) {
	headers := r.Header.Values(xmidtHeader)
	if len(headers) != 0 {
		l.metrics.TokenHeaderUsed.inc(xmidtHeader)
	} else {
		headers = r.Header.Values(webpaHeader)
		l.metrics.TokenHeaderUsed.inc(webpaHeader)
	}

	if len(headers) == 0 {
		l.metrics.TokenOutcome.incNoTokenHeader()
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
			l.metrics.TokenOutcome.incInvalidHeaderFormat()
			return nil, ErrInvalidAuth
		}

		alg := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		if alg == "" || val == "" {
			l.metrics.TokenOutcome.incInvalidHeaderFormat()
			return nil, ErrInvalidAuth
		}

		choices[alg] = val
		list = append(list, alg)
	}

	l.metrics.TokenAlgorithms.inc(list)

	best, err := l.best(list)
	if err != nil {
		l.metrics.TokenOutcome.incAlgorithmNotFound()
		return nil, err
	}

	l.metrics.TokenAlgorithmUsed.inc(best)
	l.metrics.TokenOutcome.incValid()

	return NewToken(best, choices[best]), nil
}

// Authorize validates that the request body matches the hash and secret provided
// in the token.
func (l *Listener) Authorize(r *http.Request, t Token) error {
	secret, err := hex.DecodeString(t.Principal())
	if err != nil {
		l.metrics.Authorization.incInvalidSignature()
		return fmt.Errorf("%w: unable to decode signature", ErrInput)
	}

	if r.Body == nil {
		l.metrics.Authorization.incEmptyBody()
		return fmt.Errorf("%w: empty request body", ErrInput)
	}

	msg, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		l.metrics.Authorization.incUnableToReadBody()
		return fmt.Errorf("%w: unable to read request body", ErrInput)
	}

	// Reset the body so it can be read again later.
	r.Body = io.NopCloser(bytes.NewReader(msg))

	if len(msg) == 0 {
		l.metrics.Authorization.incEmptyBody()
		return fmt.Errorf("%w: empty request body", ErrInput)
	}

	hashes, err := l.getHashes(t.Type())
	if err != nil {
		return err
	}

	for _, h := range hashes {
		h.Write(msg)
		if hmac.Equal(h.Sum(nil), secret) {
			l.metrics.Authorization.incValid()
			return nil
		}
	}

	l.metrics.Authorization.incSignatureMismatch()
	return fmt.Errorf("%w: unable to validate signature", ErrInput)
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
