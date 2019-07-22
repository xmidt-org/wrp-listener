/**
 * Copyright 2019 Comcast Cable Communications Management, LLC
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
	"io/ioutil"
	"net/http"
	"time"

	"github.com/goph/emperror"
	"github.com/xmidt-org/wrp-listener"
)

const (
	DefaultContentType  = "wrp"
	DefaultDeviceRegexp = ".*"
	DefaultTimeout      = 10 * time.Second
)

// Acquirer gets an Authorization value that can be added to an http request.
// The format of the string returned should be the key, a space, and then the
// auth string.
type Acquirer interface {
	Acquire() (string, error)
}

type SecretGetter interface {
	GetSecret() (string, error)
}

type BasicRegisterer struct {
	requestTemplate webhook.W
	registrationURL string
	client          *http.Client

	acquirer     Acquirer
	secretGetter SecretGetter
}

type BasicConfig struct {
	Timeout         time.Duration
	ClientTransport http.RoundTripper
	RegistrationURL string
	Request         webhook.W
}

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

	if config.Request.Config.URL == "" {
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

	if basic.requestTemplate.Config.ContentType == "" {
		basic.requestTemplate.Config.ContentType = DefaultContentType
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

func (b *BasicRegisterer) Register() error {
	secret, err := b.secretGetter.GetSecret()
	if err != nil {
		return emperror.Wrap(err, "Failed to get secret")
	}

	b.requestTemplate.Config.Secret = secret

	marshaledBody, errMarshal := json.Marshal(&b.requestTemplate)
	if errMarshal != nil {
		return emperror.WrapWith(errMarshal, "failed to marshal")
	}

	satToken, err := b.acquirer.Acquire()
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("POST", b.registrationURL, bytes.NewBuffer(marshaledBody))
	if satToken != "" {
		req.Header.Set("Authorization", satToken)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return emperror.WrapWith(err, "failed to make http request")
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return emperror.WrapWith(err, "failed to read body")
	}

	if resp.StatusCode != 200 {
		return emperror.WrapWith(fmt.Errorf("unable to register webhook"), "received non-200 response", "code", resp.StatusCode, "body", string(respBody[:]))
	}
	return nil
}
