// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

type eventListener struct {
	l *listener.Listener
}

func (el *eventListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := el.l.Tokenize(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = el.l.Authorize(r, token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println(string(body))

	w.WriteHeader(http.StatusOK)
}

func ExampleBasicAuth() { // nolint: govet
	server := startFakeListener()
	defer server.Close()

	// Create the listener.
	r := webhook.Registration{
		Config: webhook.DeliveryConfig{
			ContentType: "application/json",
		},
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

	el := eventListener{
		l: whl,
	}

	go func() {
		err := http.ListenAndServe(":8080", &el) // nolint: gosec
		if err != nil {
			panic(err)
		}
	}()

	// Register for webhook events, using the secret "foobar" as the shared
	// secret.
	err = whl.Register("foobar")
	if err != nil {
		panic(err)
	}

	// Output:
}

func ExampleBearerAuth() { // nolint: govet
	server := startFakeListener()
	defer server.Close()

	// Create the listener.
	r := webhook.Registration{
		Config: webhook.DeliveryConfig{
			ContentType: "application/json",
		},
		Duration: webhook.CustomDuration(5 * time.Minute),
	}

	sharedSecret := strings.Split(os.Getenv("SHARED_SECRET"), ",")
	for i := range sharedSecret {
		sharedSecret[i] = strings.TrimSpace(sharedSecret[i])
	}

	url := server.URL // replace with the URL of the webhook provider
	whl, err := listener.New(&r, url,
		listener.AuthBearer(os.Getenv("BEARER_TOKEN")),
		listener.AcceptSHA1(),
		listener.AcceptedSecrets(sharedSecret...),
	)
	if err != nil {
		panic(err)
	}

	el := eventListener{
		l: whl,
	}

	go func() {
		err := http.ListenAndServe(":8080", &el) // nolint: gosec
		if err != nil {
			panic(err)
		}
	}()

	// Register for webhook events, using the secret "foobar" as the shared
	// secret.
	err = whl.Register("foobar")
	if err != nil {
		panic(err)
	}

	// Output:
}
