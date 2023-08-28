// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"reflect"
	"strconv"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
)

// Measure is a struct that holds the metrics used by the webhook listener.  The
// default is to use a no-op counter for each.
type Measure struct {
	// Registration is a metric and configuration that holds the number of times
	// the webhook registration has been attempted.
	Registration MeasureRegistration

	// RegistrationInterval is a gauge that holds the interval used to register
	// the webhook.  The value is set when the listener is started.
	//
	// With() is not called.
	RegistrationInterval metrics.Gauge `default:"discard.NewGauge()"`

	// SecretUpdated is a counter that holds the number of times the secret has
	// been updated.
	//
	// With() is not called.
	SecretUpdated metrics.Counter `default:"discard.NewCounter()"`

	// TokenOutcome is a metric and configuration that holds the number of times
	// the Tokenize() method has been called with it's outcome.
	TokenOutcome MeasureTokenOutcome

	// TokenAlgorithmUsed is a metric and configuration that holds the number of
	// times an algorithm has been selected when a request is tokenized.
	TokenAlgorithmUsed MeasureTokenAlgorithmUsed

	// TokenAlgorithms is a metric and configuration that holds the number of
	// times an algorithm has been offered to Tokenize in a request.
	TokenAlgorithms MeasureTokenAlgorithms

	// TokenHeaderUsed is a metric and configuration that holds the number of
	// times a specific header has been used to tokenize a request.
	TokenHeaderUsed MeasureTokenHeaderUsed

	// Authorization is a metric and configuration that holds the number of times
	// the authorization has been attempted.
	Authorization MeasureAuthorization
}

func (m *Measure) init() *Measure {
	if m != nil {
		setDefaults(m)
	}
	return m
}

// MeasureRegistration is a metric and configuration that holds the number of
// times the webhook registration has been attempted.  The outcome is used as
// the label.
//
// Counter.With(Label, Outcome).Add(1) is called.
type MeasureRegistration struct {
	// Label is the label used for the counter.
	Label string `default:"outcome"`

	// AuthFetchingFailure is the label used when the auth fetching fails.
	AuthFetchingFailure string `default:"failure_fetching_auth"`

	// NewRequestFailure is the label used when the registration fails to create
	NewRequestFailure string `default:"failure_create_request"`

	// RequestFailure is the label used when the registration fails to send.
	RequestFailure string `default:"failure_http_request"`

	// StatusCodePrefix is the prefix used for the status code label.
	StatusCodePrefix string `default:""`

	// Counter is the counter used to count the number of times the registration
	// has been attempted.
	Counter metrics.Counter `default:"discard.NewCounter()"`
}

func (m *MeasureRegistration) incAuthFetchingFailure() {
	m.Counter.With(m.Label, m.AuthFetchingFailure).Add(1)
}

func (m *MeasureRegistration) incNewRequestFailure() {
	m.Counter.With(m.Label, m.NewRequestFailure).Add(1)
}

func (m *MeasureRegistration) incRequestFailure() {
	m.Counter.With(m.Label, m.RequestFailure).Add(1)
}

func (m *MeasureRegistration) incStatusCode(code int) {
	s := m.StatusCodePrefix + strconv.Itoa(code)
	m.Counter.With(m.Label, s).Add(1)
}

// MeasureToken is a metric and configuration that holds the number of times the
// Tokenize() method has been called with it's outcome.  The outcome is used as
// the label.
//
// Counter.With(Label, Outcome).Add(1) is called.
type MeasureTokenOutcome struct {
	Label               string          `default:"outcome"`
	Valid               string          `default:"success"`
	NoTokenHeader       string          `default:"failure_no_token_header"`
	InvalidHeaderFormat string          `default:"failure_invalid_header_format"`
	AlgorithmNotFound   string          `default:"failure_algorithm_not_found"`
	Counter             metrics.Counter `default:"discard.NewCounter()"`
}

func (m *MeasureTokenOutcome) incNoTokenHeader() {
	m.Counter.With(m.Label, m.NoTokenHeader).Add(1)
}

func (m *MeasureTokenOutcome) incInvalidHeaderFormat() {
	m.Counter.With(m.Label, m.InvalidHeaderFormat).Add(1)
}

func (m *MeasureTokenOutcome) incAlgorithmNotFound() {
	m.Counter.With(m.Label, m.AlgorithmNotFound).Add(1)
}

func (m *MeasureTokenOutcome) incValid() {
	m.Counter.With(m.Label, m.Valid).Add(1)
}

// MeasureTokenAlgorithmUsed is a metric and configuration that holds the number
// of times an algorithm has been used to tokenize a request.  The algorithm
// is used as the label.
//
// Counter.With(Label, alg).Add(1) is called.
type MeasureTokenAlgorithmUsed struct {
	Label   string          `default:"alg"`
	Counter metrics.Counter `default:"discard.NewCounter()"`
}

func (m *MeasureTokenAlgorithmUsed) inc(alg string) {
	m.Counter.With(m.Label, alg).Add(1)
}

// MeasureTokenAlgorithms is a metric and configuration that holds the number
// of times an algorithm has been offered to Tokenize in a request.  The algorithm
// is used as the label.
//
// Counter.With(Label, alg).Add(1) is called for each offered algorithm.
type MeasureTokenAlgorithms struct {
	Label   string          `default:"alg"`
	Counter metrics.Counter `default:"discard.NewCounter()"`
}

func (m *MeasureTokenAlgorithms) inc(algs []string) {
	for _, alg := range algs {
		m.Counter.With(m.Label, alg).Add(1)
	}
}

// MeasureTokenHeaderUsed is a metric and configuration that holds the number
//
// Counter.With(Label, header).Add(1) is called for each header.
type MeasureTokenHeaderUsed struct {
	Label   string          `default:"header"`
	Counter metrics.Counter `default:"discard.NewCounter()"`
}

func (m *MeasureTokenHeaderUsed) inc(header string) {
	m.Counter.With(m.Label, header).Add(1)
}

// MeasureAuthorization is a metric and configuration that holds the number of
// times the authorization has been attempted.  The outcome is used as the label.
//
// Counter.With(Label, Outcome).Add(1) is called.
type MeasureAuthorization struct {
	Label             string          `default:"outcome"`
	Valid             string          `default:"success"`
	InvalidSignature  string          `default:"failure_invalid_signature"`
	EmptyBody         string          `default:"failure_empty_body"`
	UnableToReadBody  string          `default:"failure_unable_to_read_body"`
	SignatureMismatch string          `default:"failure_signature_mismatch"`
	Counter           metrics.Counter `default:"discard.NewCounter()"`
}

func (m *MeasureAuthorization) incInvalidSignature() {
	m.Counter.With(m.Label, m.InvalidSignature).Add(1)
}

func (m *MeasureAuthorization) incEmptyBody() {
	m.Counter.With(m.Label, m.EmptyBody).Add(1)
}

func (m *MeasureAuthorization) incUnableToReadBody() {
	m.Counter.With(m.Label, m.UnableToReadBody).Add(1)
}

func (m *MeasureAuthorization) incSignatureMismatch() {
	m.Counter.With(m.Label, m.SignatureMismatch).Add(1)
}

func (m *MeasureAuthorization) incValid() {
	m.Counter.With(m.Label, m.Valid).Add(1)
}

// setDefaults sets the default values for any fi
func setDefaults(obj any) {
	valueOf := reflect.ValueOf(obj).Elem()

	for i := 0; i < valueOf.NumField(); i++ {
		field := valueOf.Field(i)
		fieldType := valueOf.Type().Field(i)
		tag := fieldType.Tag.Get("default")

		switch field.Kind() {
		case reflect.String:
			if tag != "" && field.IsZero() {
				field.SetString(tag)
			}
			continue
		case reflect.Struct:
			setDefaults(field.Addr().Interface())
			continue
		}

		switch field.Type().String() {
		case "metrics.Counter":
			if field.IsNil() {
				if tag != "discard.NewCounter()" {
					panic("invalid default value for counter")
				}
				field.Set(reflect.ValueOf(discard.NewCounter()))
			}
		case "metrics.Gauge":
			if field.IsNil() {
				if tag != "discard.NewGauge()" {
					panic("invalid default value for gauge")
				}
				field.Set(reflect.ValueOf(discard.NewGauge()))
			}
		case "metrics.Histogram":
			if field.IsNil() {
				if tag != "discard.NewHistogram()" {
					panic("invalid default value for histogram")
				}
				field.Set(reflect.ValueOf(discard.NewHistogram()))
			}
		}
	}
}
