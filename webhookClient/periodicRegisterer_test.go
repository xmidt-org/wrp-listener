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
	"errors"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"

	// nolint:staticcheck
	"github.com/xmidt-org/webpa-common/v2/xmetrics/xmetricstest"
	webhook "github.com/xmidt-org/wrp-listener"
)

// TODO: add unit tests

func TestNewPeriodicRegisterer(t *testing.T) {
	mockAcquirer := new(MockAcquirer)
	mockSecretGetter := new(mockSecretGetter)
	registry := xmetricstest.NewProvider(nil, Metrics)
	m := NewMeasures(registry)

	basicRegisterer := BasicRegisterer{
		acquirer:        mockAcquirer,
		secretGetter:    mockSecretGetter,
		requestTemplate: webhook.W{},
		registrationURL: "random string",
	}

	logger := log.NewNopLogger()
	validInterval, _ := time.ParseDuration("10s")

	tests := []struct {
		description        string
		registerer         Registerer
		interval           time.Duration
		logger             log.Logger
		expectedRegisterer *PeriodicRegisterer
		expectedErr        error
	}{
		{
			description: "Success",
			registerer:  &basicRegisterer,
			interval:    validInterval,
			logger:      logger,
			expectedRegisterer: &PeriodicRegisterer{
				registerer:           &basicRegisterer,
				registrationInterval: validInterval,
				logger:               logger,
				measures:             m,
			},
			expectedErr: nil,
		},
		{
			description: "Success with Default Logger",
			registerer:  &basicRegisterer,
			interval:    validInterval,
			expectedRegisterer: &PeriodicRegisterer{
				registerer:           &basicRegisterer,
				registrationInterval: validInterval,
				logger:               defaultLogger,
				measures:             m,
			},
			expectedErr: nil,
		},
		{
			description:        "0 interval",
			registerer:         &basicRegisterer,
			interval:           0,
			expectedRegisterer: nil,
			expectedErr:        errors.New("interval cannot be 0"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			m := NewMeasures(registry)
			pr, err := NewPeriodicRegisterer(tc.registerer, tc.interval, tc.logger, m)
			if pr != nil {
				// make sure shutdown channel is created
				assert.NotNil(pr.shutdown)
				tc.expectedRegisterer.shutdown = pr.shutdown
			}

			assert.Equal(tc.expectedRegisterer, pr)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}
