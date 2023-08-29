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
	whl, err := New(&webhook.Registration{
		Address: server.URL,
		Events: []string{
			"foo",
		},
		Config: webhook.DeliveryConfig{
			Secret: "secret1",
		},
		Duration: webhook.CustomDuration(5 * time.Minute),
	},
		Interval(1*time.Millisecond),
		AuthBasic("user", "pass"),
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
	whl, err := New(&webhook.Registration{
		Address: server.URL,
		Events: []string{
			"foo",
		},
		Config: webhook.DeliveryConfig{
			Secret: "secret1",
		},
		Duration: webhook.CustomDuration(5 * time.Minute),
	},
		Once(),
	)
	require.NotNil(whl)
	require.NoError(err)

	// Register the webhook.
	err = whl.Register()
	assert.NoError(err)

	// Re-register because it could happen.
	err = whl.Register()
	assert.NoError(err)

	// Wait a bit then roll the secret..
	time.Sleep(time.Millisecond)
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
	time.Sleep(time.Millisecond)
	m.Lock()
	expectSecret = []string{"secret5"}
	m.Unlock()

	// Wait a bit then unregister.
	time.Sleep(time.Millisecond)
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

	// Create the listener.
	whl, err := New(&webhook.Registration{
		Address: server.URL,
		Events: []string{
			"foo",
		},
		Config: webhook.DeliveryConfig{
			Secret: "secret1",
		},
		Duration: webhook.CustomDuration(5 * time.Minute),
	},
		Once(),
	)

	require.NotNil(whl)
	require.NoError(err)

	// Register the webhook.
	err = whl.Register()
	assert.ErrorIs(err, ErrRegistrationFailed)
}

func TestFailedAuthCheck(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Create the listener.
	whl, err := New(&webhook.Registration{
		Events: []string{
			"foo",
		},
		Config: webhook.DeliveryConfig{
			Secret: "secret1",
		},
		Duration: webhook.CustomDuration(5 * time.Minute),
	},
		AuthBearerFunc(func() (string, error) {
			return "", fmt.Errorf("nope")
		}),
	)

	require.NotNil(whl)
	require.NoError(err)

	// Register the webhook.
	err = whl.Register()
	assert.ErrorIs(err, ErrRegistrationNotAttempted)
}

func TestFailedNewRequest(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Create the listener.
	whl, err := New(&webhook.Registration{
		Address: "//invalid::localhost/:99999",
		Events: []string{
			"foo",
		},
		Config: webhook.DeliveryConfig{
			Secret: "secret1",
		},
		Duration: webhook.CustomDuration(5 * time.Minute),
	})

	require.NotNil(whl)
	require.NoError(err)

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

	// Create the listener.
	whl, err := New(&webhook.Registration{
		Address: server.URL,
		Events: []string{
			"foo",
		},
		Config: webhook.DeliveryConfig{
			Secret: "secret1",
		},
		Duration: webhook.CustomDuration(5 * time.Minute),
	},
		HTTPClient(&http.Client{Timeout: 1 * time.Millisecond}),
		Once(),
	)

	require.NotNil(whl)
	require.NoError(err)

	// Register the webhook.
	err = whl.Register()
	assert.ErrorIs(err, ErrRegistrationFailed)
}
