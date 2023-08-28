// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/stretchr/testify/assert"
)

func TestMeasure_init(t *testing.T) {
	tests := []struct {
		description string
		in          Measure
		expected    Measure
	}{
		{
			description: "metrics test",
			in:          Measure{},
			expected: Measure{
				Registration: MeasureRegistration{
					Label:               "outcome",
					AuthFetchingFailure: "failure_fetching_auth",
					NewRequestFailure:   "failure_create_request",
					RequestFailure:      "failure_http_request",
					Counter:             discard.NewCounter(),
				},
				RegistrationInterval: discard.NewGauge(),
				SecretUpdated:        discard.NewCounter(),
				TokenOutcome: MeasureTokenOutcome{
					Label:               "outcome",
					Valid:               "success",
					NoTokenHeader:       "failure_no_token_header",
					InvalidHeaderFormat: "failure_invalid_header_format",
					AlgorithmNotFound:   "failure_algorithm_not_found",
					Counter:             discard.NewCounter(),
				},
				TokenAlgorithmUsed: MeasureTokenAlgorithmUsed{
					Label:   "alg",
					Counter: discard.NewCounter(),
				},
				TokenAlgorithms: MeasureTokenAlgorithms{
					Label:   "alg",
					Counter: discard.NewCounter(),
				},
				TokenHeaderUsed: MeasureTokenHeaderUsed{
					Label:   "header",
					Counter: discard.NewCounter(),
				},
				Authorization: MeasureAuthorization{
					Label:             "outcome",
					Valid:             "success",
					InvalidSignature:  "failure_invalid_signature",
					EmptyBody:         "failure_empty_body",
					UnableToReadBody:  "failure_unable_to_read_body",
					SignatureMismatch: "failure_signature_mismatch",
					Counter:           discard.NewCounter(),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			in := tc.in
			in.init()
			assert.Equal(tc.expected, in)
			//pp.Println(tc.in)
		})
	}
}

func Test_setDefaults(t *testing.T) {
	type sub struct {
		Foo string `default:"bar"`
	}

	type simple struct {
		Label     string            `default:"label"`
		Counter   metrics.Counter   `default:"discard.NewCounter()"`
		Gauge     metrics.Gauge     `default:"discard.NewGauge()"`
		Histogram metrics.Histogram `default:"discard.NewHistogram()"`
		Sub       sub
	}

	type invalidCounter struct {
		Counter metrics.Counter `default:"invalid"`
	}

	type invalidGauge struct {
		Gauge metrics.Gauge `default:"invalid"`
	}

	type invalidHistogram struct {
		Histogram metrics.Histogram `default:"invalid"`
	}

	tests := []struct {
		description string
		in          any
		expected    any
	}{
		{
			description: "basic test",
			in:          &simple{},
			expected: &simple{
				Label:     "label",
				Counter:   discard.NewCounter(),
				Gauge:     discard.NewGauge(),
				Histogram: discard.NewHistogram(),
				Sub: sub{
					Foo: "bar",
				},
			},
		}, {
			description: "basic test",
			in: &simple{
				Label:     "foo",
				Counter:   discard.NewCounter(),
				Gauge:     discard.NewGauge(),
				Histogram: discard.NewHistogram(),
			},
			expected: &simple{
				Label:     "foo",
				Counter:   discard.NewCounter(),
				Gauge:     discard.NewGauge(),
				Histogram: discard.NewHistogram(),
				Sub: sub{
					Foo: "bar",
				},
			},
		}, {
			description: "invalid counter test",
			in:          &invalidCounter{},
		}, {
			description: "invalid gauge test",
			in:          &invalidGauge{},
		}, {
			description: "invalid histogram test",
			in:          &invalidHistogram{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			in := tc.in
			if tc.expected == nil {
				assert.Panics(func() { setDefaults(in) })
				return
			}

			setDefaults(in)
			assert.Equal(tc.expected, in)
			//pp.Println(tc.in)
		})
	}
}
