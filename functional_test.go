// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/webhook-schema"
)

func TestNormalUsage(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var m sync.Mutex

	expectSecret := []string{"secret1"}

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				assert.NoError(err)
				r.Body.Close()

				var reg webhook.Registration
				err = json.Unmarshal(body, &reg)
				assert.NoError(err)

				found := false
				m.Lock()
				defer m.Unlock()
				for _, s := range expectSecret {
					if s == reg.Config.Secret {
						found = true
						break
					}
				}

				assert.True(found)

				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()

	// Create the listener.
	whl, err := New(
		&webhook.Registration{
			Events: []string{
				"foo",
			},
			Config: webhook.DeliveryConfig{
				Secret: "secret1",
			},
			Duration: webhook.CustomDuration(5 * time.Minute),
		},
		server.URL,
		Interval(1*time.Millisecond),
	)
	require.NotNil(whl)
	require.NoError(err)

	// Register the webhook before has started
	err = whl.Register("secret1")
	assert.NoError(err)

	err = whl.Register("secret1")
	assert.NoError(err)

	// Register the webhook.
	err = whl.Register()
	assert.NoError(err)

	// Re-register because it could happen.
	err = whl.Register()
	assert.NoError(err)

	// Wait a bit then roll the secret..
	time.Sleep(time.Millisecond)
	m.Lock()
	expectSecret = append(expectSecret, "secret2")
	m.Unlock()

	err = whl.Register("secret2")
	assert.NoError(err)

	// Wait a bit then remove the prior secret from the list of accepted secrets.
	time.Sleep(time.Millisecond)
	m.Lock()
	expectSecret = []string{"secret2"}
	m.Unlock()

	// Wait a bit then unregister.
	time.Sleep(time.Millisecond)
	whl.Stop()

	// Re-stop because it could happen.
	whl.Stop()
}

func TestSingleShotUsage(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var m sync.Mutex

	expectSecret := []string{"secret1"}

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				assert.NoError(err)
				r.Body.Close()

				var reg webhook.Registration
				err = json.Unmarshal(body, &reg)
				assert.NoError(err)

				found := false
				m.Lock()
				for _, s := range expectSecret {
					if s == reg.Config.Secret {
						found = true
						break
					}
				}
				m.Unlock()

				assert.True(found)

				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()

	eventListener := testListener{
		re: func(e EventRegistration) {
			assert.Equal(http.StatusOK, e.StatusCode)
			assert.NotZero(e.At)
			assert.NotZero(e.Duration)
			assert.NoError(e.Err)
		},
	}

	// Create the listener.
	whl, err := New(
		&webhook.Registration{
			Events: []string{
				"foo",
			},
			Config: webhook.DeliveryConfig{
				Secret: "secret1",
			},
			Duration: webhook.CustomDuration(5 * time.Minute),
		},
		server.URL,
		Once(),
	)
	require.NotNil(whl)
	require.NoError(err)

	cancel := whl.AddRegistrationEventListener(&eventListener)
	assert.NotNil(cancel)

	// Register the webhook.
	err = whl.Register()
	assert.NoError(err)

	// Re-register because it could happen.
	err = whl.Register()
	assert.NoError(err)

	// Wait a bit then roll the secret..
	time.Sleep(10 * time.Millisecond)
	m.Lock()
	expectSecret = append(expectSecret, "secret2", "secret3", "secret4")
	m.Unlock()

	err = whl.Register("secret2")
	assert.NoError(err)

	err = whl.Register("secret3")
	assert.NoError(err)

	err = whl.Register("secret4")
	assert.NoError(err)

	// Wait a bit then remove the prior secret from the list of accepted secrets.
	time.Sleep(10 * time.Millisecond)
	m.Lock()
	expectSecret = []string{"secret5"}
	m.Unlock()

	// Wait a bit then unregister.
	time.Sleep(10 * time.Millisecond)
	whl.Stop()

	// Re-stop because it could happen.
	whl.Stop()
}

func TestFailedHTTPCall(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
		),
	)
	defer server.Close()

	eventListener := testListener{
		re: func(e EventRegistration) {
			assert.Equal(http.StatusBadRequest, e.StatusCode)
			assert.NotZero(e.At)
			assert.NotZero(e.Duration)
			assert.ErrorIs(e.Err, ErrRegistrationFailed)
		},
	}

	// Create the listener.
	whl, err := New(
		&webhook.Registration{
			Events: []string{
				"foo",
			},
			Config: webhook.DeliveryConfig{
				Secret: "secret1",
			},
			Duration: webhook.CustomDuration(5 * time.Minute),
		},
		server.URL,
		Once(),
	)

	_ = whl.AddRegistrationEventListener(&eventListener)

	require.NotNil(whl)
	require.NoError(err)

	// Register the webhook.
	err = whl.Register()
	assert.ErrorIs(err, ErrRegistrationFailed)
}

func TestFailedAuthCheck(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	eventListener := testListener{
		re: func(e EventRegistration) {
			assert.Zero(e.StatusCode)
			assert.Zero(e.At)
			assert.Zero(e.Duration)
			assert.ErrorIs(e.Err, ErrRegistrationNotAttempted)
			assert.ErrorIs(e.Err, ErrDecoratorFailed)
		},
	}

	// Create the listener.
	whl, err := New(
		&webhook.Registration{
			Events: []string{
				"foo",
			},
			Config: webhook.DeliveryConfig{
				Secret: "secret1",
			},
			Duration: webhook.CustomDuration(5 * time.Minute),
		},
		"http://example.com",
		DecorateRequest(
			DecoratorFunc(func(r *http.Request) error {
				return fmt.Errorf("nope")
			}),
		),
	)

	require.NotNil(whl)
	require.NoError(err)

	cancel := whl.AddRegistrationEventListener(&eventListener)
	assert.NotNil(cancel)

	// Register the webhook.
	err = whl.Register()
	assert.ErrorIs(err, ErrRegistrationNotAttempted)
}

func TestFailedNewRequest(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	eventListener := testListener{
		re: func(e EventRegistration) {
			assert.Zero(e.StatusCode)
			assert.Zero(e.At)
			assert.Zero(e.Duration)
			assert.ErrorIs(e.Err, ErrRegistrationNotAttempted)
			assert.ErrorIs(e.Err, ErrNewRequestFailed)
		},
	}

	// Create the listener.
	whl, err := New(
		&webhook.Registration{
			Events: []string{
				"foo",
			},
			Config: webhook.DeliveryConfig{
				Secret: "secret1",
			},
			Duration: webhook.CustomDuration(5 * time.Minute),
		},
		"//invalid::localhost/:99999",
	)

	require.NotNil(whl)
	require.NoError(err)

	cancel := whl.AddRegistrationEventListener(&eventListener)
	assert.NotNil(cancel)

	// Register the webhook.
	err = whl.Register()
	assert.ErrorIs(err, ErrRegistrationNotAttempted)
}

func TestCancelListener(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	eventListener := testListener{
		re: func(e EventRegistration) {
			assert.Fail("should not have been called")
		},
	}

	// Create the listener.
	whl, err := New(
		&webhook.Registration{
			Events: []string{
				"foo",
			},
			Config: webhook.DeliveryConfig{
				Secret: "secret1",
			},
			Duration: webhook.CustomDuration(5 * time.Minute),
		},
		"//invalid::localhost/:99999",
	)

	require.NotNil(whl)
	require.NoError(err)

	cancel := whl.AddRegistrationEventListener(&eventListener)
	assert.NotNil(cancel)
	cancel()

	// Register the webhook.
	err = whl.Register()
	assert.ErrorIs(err, ErrRegistrationNotAttempted)
}

func TestFailedConnect(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(10 * time.Millisecond)
			},
		),
	)
	defer server.Close()

	eventListener := testListener{
		re: func(e EventRegistration) {
			assert.Zero(e.StatusCode)
			assert.NotZero(e.At)
			assert.NotZero(e.Duration)
			assert.ErrorIs(e.Err, ErrRegistrationFailed)
		},
	}

	// Create the listener.
	whl, err := New(
		&webhook.Registration{
			Events: []string{
				"foo",
			},
			Config: webhook.DeliveryConfig{
				Secret: "secret1",
			},
			Duration: webhook.CustomDuration(5 * time.Minute),
		},
		server.URL,
		HTTPClient(&http.Client{Timeout: 1 * time.Millisecond}),
		Once(),
	)

	require.NotNil(whl)
	require.NoError(err)

	cancel := whl.AddRegistrationEventListener(&eventListener)
	assert.NotNil(cancel)

	// Register the webhook.
	err = whl.Register()
	assert.ErrorIs(err, ErrRegistrationFailed)
}

func TestFailsAfterABit(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var m sync.Mutex
	var count int

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.ReadAll(r.Body)
				r.Body.Close()

				m.Lock()
				if count == 0 {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
				count++
				m.Unlock()
			},
		),
	)
	defer server.Close()

	eventListener := testListener{
		re: func(e EventRegistration) {
			assert.NotZero(e.StatusCode)
			assert.NotZero(e.At)
			assert.NotZero(e.Duration)
			if e.StatusCode == http.StatusOK {
				assert.NoError(e.Err)
				return
			}
			assert.ErrorIs(e.Err, ErrRegistrationFailed)
		},
	}

	// Create the listener.
	whl, err := New(
		&webhook.Registration{
			Events: []string{
				"foo",
			},
			Config: webhook.DeliveryConfig{
				Secret: "secret1",
			},
			Duration: webhook.CustomDuration(5 * time.Minute),
		},
		server.URL,
		Interval(1*time.Millisecond),
	)
	require.NotNil(whl)
	require.NoError(err)

	cancel := whl.AddRegistrationEventListener(&eventListener)
	assert.NotNil(cancel)

	// Register the webhook before has started
	err = whl.Register()
	assert.NoError(err)

	// Wait a bit then roll the secret..
	time.Sleep(10 * time.Millisecond)

	whl.Stop()
}
