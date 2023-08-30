// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/xmidt-org/webhook-schema"
	listener "github.com/xmidt-org/wrp-listener"
	"go.uber.org/zap"
)

type eventListener struct {
	l *listener.Listener
}

func (el *eventListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got a request")
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

func main() {
	receiverURL := strings.TrimSpace(os.Getenv("WEBHOOK_TARGET"))
	webhookURL := strings.TrimSpace(os.Getenv("WEBHOOK_URL"))
	localAddress := strings.TrimSpace(os.Getenv("WEBHOOK_LISTEN_ADDR"))
	certFile := strings.TrimSpace(os.Getenv("WEBHOOK_LISTEN_CERT_FILE"))
	keyFile := strings.TrimSpace(os.Getenv("WEBHOOK_LISTEN_KEY_FILE"))

	useTLS := false
	if certFile != "" && keyFile != "" {
		useTLS = true
	}

	fmt.Println("WEBHOOK_TARGET          : ", receiverURL)
	fmt.Println("WEBHOOK_URL             : ", webhookURL)
	fmt.Println("WEBHOOK_LISTEN_ADDR     : ", localAddress)
	fmt.Println("WEBHOOK_LISTEN_CERT_FILE: ", certFile)
	fmt.Println("WEBHOOK_LISTEN_KEY_FILE : ", keyFile)
	fmt.Printf("                 use TLS: %t\n", useTLS)

	// Create the listener.
	r := webhook.Registration{
		Config: webhook.DeliveryConfig{
			ReceiverURL: receiverURL,
			ContentType: "application/json",
		},
		Events: []string{
			"device-status:*",
		},
		Duration: webhook.CustomDuration(5 * time.Minute),
	}

	sharedSecrets := strings.Split(os.Getenv("WEBHOOK_SHARED_SECRETS"), ",")
	for i := range sharedSecrets {
		sharedSecrets[i] = strings.TrimSpace(sharedSecrets[i])
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	whl, err := listener.New(&r, webhookURL,
		listener.AuthBearer(os.Getenv("WEBHOOK_BEARER_TOKEN")),
		listener.AcceptSHA1(),
		listener.Logger(logger),
		listener.Once(),
		listener.AcceptedSecrets(sharedSecrets...),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(whl.String())

	el := eventListener{
		l: whl,
	}

	go func() {
		if useTLS {
			err := http.ListenAndServeTLS(localAddress, certFile, keyFile, &el) // nolint: gosec
			if err != nil {
				panic(err)
			}
		} else {
			err := http.ListenAndServe(localAddress, &el) // nolint: gosec
			if err != nil {
				panic(err)
			}
		}
	}()

	// Register for webhook events, using the secret "foobar" as the shared
	// secret.
	err = whl.Register("foobar")
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(1 * time.Minute)
	}
}
