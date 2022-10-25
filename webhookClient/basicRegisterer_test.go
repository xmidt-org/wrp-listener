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
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	webhook "github.com/xmidt-org/wrp-listener"
)

func TestNewBasicRegisterer(t *testing.T) {
	mockAcquirer := new(MockAcquirer)
	mockSecretGetter := new(mockSecretGetter)
	mockTransport := new(MockRoundTripper)
	tests := []struct {
		description        string
		acquirer           Acquirer
		secret             SecretGetter
		config             BasicConfig
		expectedRegisterer *BasicRegisterer
		expectedErr        error
	}{
		{
			description: "Success",
			acquirer:    mockAcquirer,
			secret:      mockSecretGetter,
			config: BasicConfig{
				RegistrationURL: "/r",
				Timeout:         5 * time.Minute,
				ClientTransport: mockTransport,
				Request: webhook.W{
					Config: webhook.Config{
						URL:         "/",
						ContentType: "text/json",
					},
					Events: []string{""},
					Matcher: webhook.Matcher{
						DeviceID: []string{"mac:1234.*"},
					},
				},
			},
			expectedRegisterer: &BasicRegisterer{
				acquirer:     mockAcquirer,
				secretGetter: mockSecretGetter,
				client: &http.Client{
					Timeout:   5 * time.Minute,
					Transport: mockTransport,
				},
				registrationURL: "/r",
				requestTemplate: webhook.W{
					Config: webhook.Config{
						URL:         "/",
						ContentType: "text/json",
					},
					Events: []string{""},
					Matcher: webhook.Matcher{
						DeviceID: []string{"mac:1234.*"},
					},
				},
			},
			expectedErr: nil,
		},
		{
			description: "Success With Defaults",
			acquirer:    mockAcquirer,
			secret:      mockSecretGetter,
			config: BasicConfig{
				RegistrationURL: "/r",
				Request: webhook.W{
					Config: webhook.Config{
						URL: "/",
					},
					Events: []string{""},
				},
			},
			expectedRegisterer: &BasicRegisterer{
				acquirer:     mockAcquirer,
				secretGetter: mockSecretGetter,
				client: &http.Client{
					Timeout: DefaultTimeout,
				},
				registrationURL: "/r",
				requestTemplate: webhook.W{
					Config: webhook.Config{
						URL:         "/",
						ContentType: DefaultContentType,
					},
					Events: []string{""},
					Matcher: webhook.Matcher{
						DeviceID: []string{DefaultDeviceRegexp},
					},
				},
			},
			expectedErr: nil,
		},
		{
			description: "Nil Acquirer Error",
			expectedErr: errors.New("nil Acquirer"),
		},
		{
			description: "Nil SecretGetter Error",
			acquirer:    mockAcquirer,
			expectedErr: errors.New("nil SecretGetter"),
		},
		{
			description: "Empty Registration URL Error",
			acquirer:    mockAcquirer,
			secret:      mockSecretGetter,
			expectedErr: errors.New("invalid registration URL"),
		},
		{
			description: "Empty Request Config URL Error",
			acquirer:    mockAcquirer,
			secret:      mockSecretGetter,
			config: BasicConfig{
				RegistrationURL: "/",
			},
			expectedErr: errors.New("invalid webhook config URL"),
		},
		{
			description: "Empty Request Events Error",
			acquirer:    mockAcquirer,
			secret:      mockSecretGetter,
			config: BasicConfig{
				RegistrationURL: "/r",
				Request: webhook.W{
					Config: webhook.Config{
						URL: "/",
					},
				},
			},
			expectedErr: errors.New("need at least one regular expression"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			br, err := NewBasicRegisterer(tc.acquirer, tc.secret, tc.config)
			assert.Equal(tc.expectedRegisterer, br)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestBasicRegister(t *testing.T) {
	goodResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`OK`)),
		Header:     make(http.Header),
	}
	badResponse := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewBufferString(`OK`)),
		Header:     make(http.Header),
	}

	testErr := errors.New("test error")
	acquireErr := errors.New("test acquire error")

	tests := []struct {
		description        string
		getSecretErr       error
		acquireCalled      bool
		acquireErr         error
		request            bool
		requestResponse    *http.Response
		requestResponseErr error
		expectedErr        error
		expectedReason     ReasonCode
	}{
		{
			description:     "Success",
			acquireCalled:   true,
			request:         true,
			requestResponse: goodResponse,
			expectedErr:     nil,
		},
		{
			description:    "Get Secret Error",
			getSecretErr:   testErr,
			expectedErr:    ErrGetSecretFail,
			expectedReason: GetSecretFail,
		},
		{
			description:    "Acquire Error",
			acquireCalled:  true,
			acquireErr:     acquireErr,
			expectedErr:    acquireErr,
			expectedReason: AcquireJWTFail,
		},
		{
			description:        "Do Error",
			acquireCalled:      true,
			request:            true,
			requestResponse:    goodResponse,
			requestResponseErr: testErr,
			expectedErr:        ErrDoFail,
			expectedReason:     DoRequestFail,
		},
		{
			description:     "Bad Response Error",
			acquireCalled:   true,
			request:         true,
			requestResponse: badResponse,
			expectedErr:     ErrNon200Resp,
			expectedReason:  Non200Response,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockTransport := new(MockRoundTripper)
			if tc.request {
				mockTransport.On("RoundTrip", mock.Anything, mock.Anything).Return(tc.requestResponse, tc.requestResponseErr).Once()
			}
			client := &http.Client{
				Transport: mockTransport,
			}

			mockAcquirer := new(MockAcquirer)
			if tc.acquireCalled {
				mockAcquirer.On("Acquire").Return("testtoken", tc.acquireErr).Once()
			}
			mockSecretGetter := new(mockSecretGetter)
			mockSecretGetter.On("GetSecret").Return("testsecret", tc.getSecretErr).Once()
			wh := &BasicRegisterer{
				acquirer:     mockAcquirer,
				client:       client,
				secretGetter: mockSecretGetter,
			}
			err := wh.Register()
			mockTransport.AssertExpectations(t)
			mockAcquirer.AssertExpectations(t)
			mockSecretGetter.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
			if tc.expectedErr != nil {
				assert.Equal(tc.expectedReason, GetReasonCode(err))
			}
		})
	}
}
