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

package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/xmidt-org/bascule/basculehttp"
	webhook "github.com/xmidt-org/wrp-listener"
	"github.com/xmidt-org/wrp-listener/hashTokenFactory"
	"net/http"
	"os"
	"time"

	"github.com/justinas/alice"

	"github.com/xmidt-org/bascule/acquire"
	"github.com/xmidt-org/wrp-listener/secret"
	"github.com/xmidt-org/wrp-listener/webhookClient"
)

func main() {

	// use constant secret for hash
	secretGetter := secretGetter.NewConstantSecret("secret1234")

	// set up the middleware
	htf, err := hashTokenFactory.New("sha1", sha1.New, secretGetter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup hash token factory: %v\n", err.Error())
		os.Exit(1)
	}
	authConstructor := basculehttp.NewConstructor(
		basculehttp.WithTokenFactory("sha1", htf),
		basculehttp.WithHeaderName("X-Webpa-Signature"),
		basculehttp.WithHeaderDelimiter("="),
	)
	handler := alice.New(authConstructor)

	// set up the registerer
	basicConfig := webhookClient.BasicConfig{
		Timeout:         5 * time.Second,
		RegistrationURL: "http://tr1d1um:6100/api/v3/hook",
		Request: webhook.W{
			Config: webhook.Config{
				URL: "http://listener-example:7100/events",
				Secret: "secret1234",
			},
			Events:   []string{".*"},
			Duration: time.Minute * 1,
		},
	}

	// This Basic Auth credentials intended to be used for local testing purposes.
	// Change this.
	acquirer, err := acquire.NewFixedAuthAcquirer("Basic dXNlcjpwYXNz")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create fixed auth: %v\n", err.Error())
		os.Exit(1)
	}
	registerer, err := webhookClient.NewBasicRegisterer(acquirer, secretGetter, basicConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup registerer: %v\n", err.Error())
		os.Exit(1)
	}
	periodicRegisterer, err := webhookClient.NewPeriodicRegisterer(registerer, 55*time.Second, log.NewLogfmtLogger(os.Stdout), webhookClient.NewMeasures(provider.NewDiscardProvider()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup periodic registerer: %v\n", err.Error())
		os.Exit(1)
	}
	// start the registerer
	periodicRegisterer.Start()

	// start listening
	http.Handle("/events", handler.ThenFunc(return200))
	err = http.ListenAndServe(":7100", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error serving http requests: %v\n", err.Error())
		os.Exit(1)
	}

}

func return200(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received http request")
	w.WriteHeader(http.StatusOK)
}
