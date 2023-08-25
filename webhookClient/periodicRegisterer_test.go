// SPDX-FileCopyrightText: 2019 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package webhookClient

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

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

	logger := zap.NewNop()
	validInterval, _ := time.ParseDuration("10s")

	tests := []struct {
		description        string
		registerer         Registerer
		interval           time.Duration
		logger             *zap.Logger
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
				logger:               logger,
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
