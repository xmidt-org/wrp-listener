/**
 * Copyright 2020 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package webhookClient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	webhook "github.com/xmidt-org/wrp-listener"
)

const (
	DefaultContentType  = "wrp"
	DefaultDeviceRegexp = ".*"
	DefaultTimeout      = 10 * time.Second
)

var (
	ErrGetSecretFail = errors.New("failed to get secret")
	ErrMarshalFail   = errors.New("failed to marshal request")
	ErrDoFail        = errors.New("failed to make http request")
	ErrReadFail      = errors.New("failed to read body")
	ErrNon200Resp    = errors.New("received non-200 response")
)

// Acquirer gets an Authorization value that can be added to an http request.
// The format of the string returned should be the key, a space, and then the
// auth string.
type Acquirer interface {
	Acquire() (string, error)
}

// SecretGetter gets the secret to use when hashing.  If getting the secret is
// unsuccessful, an error can be returned.
type SecretGetter interface {
	GetSecret() (string, error)
}

// BasicRegisterer sends POST requests to register at the webhook URL provided.
type BasicRegisterer struct {
	requestTemplate webhook.Config
	registrationURL string
	client          *http.Client

	acquirer     Acquirer
	secretGetter SecretGetter
}

// BasicConfig holds the configuration options for setting up a
// BasicRegisterer.
type BasicConfig struct {
	Timeout         time.Duration
	ClientTransport http.RoundTripper
	RegistrationURL string
	Request         webhook.Config
}

// NewBasicRegisterer returns a basic registerer set up with the configuration
// given.  If the acquirer or secretGetter are nil or certain configurations
// are empty, an error will be returned.  Otherwise, some config values are
// set to defaults if they are invalid and a basic registerer is returned.
func NewBasicRegisterer(acquirer Acquirer, secret SecretGetter, config BasicConfig) (*BasicRegisterer, error) {
	if acquirer == nil {
		return nil, errors.New("nil Acquirer")
	}

	if secret == nil {
		return nil, errors.New("nil SecretGetter")
	}

	if config.RegistrationURL == "" {
		return nil, errors.New("invalid registration URL")
	}

	if config.Request.URL == "" {
		return nil, errors.New("invalid webhook config URL")
	}

	if len(config.Request.Events) == 0 {
		return nil, errors.New("need at least one regular expression to match to events")
	}

	basic := BasicRegisterer{
		acquirer:        acquirer,
		secretGetter:    secret,
		registrationURL: config.RegistrationURL,
		requestTemplate: config.Request,
	}

	if basic.requestTemplate.ContentType == "" {
		basic.requestTemplate.ContentType = DefaultContentType
	}

	if len(basic.requestTemplate.Matcher.DeviceID) == 0 {
		basic.requestTemplate.Matcher.DeviceID = []string{DefaultDeviceRegexp}
	}

	httpTimeout := config.Timeout
	if httpTimeout <= 0 {
		httpTimeout = DefaultTimeout
	}

	basic.client = &http.Client{
		Timeout: httpTimeout,
	}

	if config.ClientTransport != nil {
		basic.client.Transport = config.ClientTransport
	}

	return &basic, nil
}

// Register registers to the webhook using the information the basic registerer
// has.
func (b *BasicRegisterer) Register() error {
	secret, err := b.secretGetter.GetSecret()
	if err != nil {
		return errWithReason{
			err:    fmt.Errorf("%w: %v", ErrGetSecretFail, err),
			reason: GetSecretFail,
		}
	}
	b.requestTemplate.Secret = secret
	marshaledBody, errMarshal := json.Marshal(&b.requestTemplate)
	if errMarshal != nil {
		return errWithReason{
			err:    fmt.Errorf("%w: %v", ErrMarshalFail, errMarshal),
			reason: MarshalRequestFail,
		}
	}

	jwtToken, err := b.acquirer.Acquire()
	if err != nil {
		return errWithReason{
			err:    err,
			reason: AcquireJWTFail,
		}
	}

	req, err := http.NewRequest("POST", b.registrationURL, bytes.NewBuffer(marshaledBody))
	if err != nil {
		return errWithReason{
			err:    err,
			reason: CreateRequestFail,
		}
	}

	if jwtToken != "" {
		req.Header.Set("Authorization", jwtToken)
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return errWithReason{
			err:    fmt.Errorf("%w: %v", ErrDoFail, err),
			reason: DoRequestFail,
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errWithReason{
			err:    fmt.Errorf("%w: %v", ErrReadFail, err),
			reason: ReadBodyFail,
		}
	}

	if resp.StatusCode != 200 {
		return errWithReason{
			err: fmt.Errorf("%w: %v, body: %v", ErrNon200Resp,
				resp.StatusCode, string(respBody[:])),
			reason: Non200Response,
		}
	}
	return nil
}
