// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/xmidt-org/webhook-schema"
	listener "github.com/xmidt-org/wrp-listener"
)

func startFakeListener() *httptest.Server {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				r.Body.Close()

				var reg webhook.Registration
				_ = json.Unmarshal(body, &reg)

				w.WriteHeader(http.StatusOK)
			},
		),
	)

	return server
}

func ExampleBasicAuth() { // nolint: govet
	server := startFakeListener()
	defer server.Close()

	// Create the listener.
	r := webhook.Registration{
		Duration: webhook.CustomDuration(5 * time.Minute),
	}

	url := server.URL // replace with the URL of the webhook provider
	whl, err := listener.New(&r, url,
		listener.AuthBasic("username", "password"),
		listener.AcceptSHA1(),
		listener.AcceptedSecrets("foobar", "carport"),
	)
	if err != nil {
		panic(err)
	}

	// Register for webhook events, using the secret "foobar" as the shared
	// secret.
	err = whl.Register("foobar")
	if err != nil {
		panic(err)
	}

	// Output:
}
