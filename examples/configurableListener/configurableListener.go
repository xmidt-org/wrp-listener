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
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"

	"github.com/justinas/alice"

	"github.com/xmidt-org/bascule/acquire"
	"github.com/xmidt-org/bascule/basculehttp"
	"github.com/xmidt-org/wrp-listener"
	"github.com/xmidt-org/wrp-listener/hashTokenFactory"
	"github.com/xmidt-org/wrp-listener/secret"
	"github.com/xmidt-org/wrp-listener/webhookClient"
)

const (
	applicationName = "configurableListener"
)

type Config struct {
	AuthHeader                  string
	AuthDelimiter               string
	WebhookRequest              webhook.W
	WebhookRegistrationURL      string
	WebhookTimeout              time.Duration
	WebhookRegistrationInterval time.Duration
	Port                        string
	Endpoint                    string
	ResponseCode                int
}

func main() {
	// load configuration with viper
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName(applicationName)
	err := v.ReadInConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read in viper config: %v\n", err.Error())
		os.Exit(1)
	}
	config := new(Config)
	err = v.Unmarshal(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal config: %v\n", err.Error())
		os.Exit(1)
	}

	// use constant secret for hash
	secretGetter := secretGetter.NewConstantSecret(config.WebhookRequest.Config.Secret)

	// set up the middleware
	htf, err := hashTokenFactory.New("Sha1", sha1.New, secretGetter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup hash token factory: %v\n", err.Error())
		os.Exit(1)
	}
	authConstructor := basculehttp.NewConstructor(
		basculehttp.WithTokenFactory("Sha1", htf),
		basculehttp.WithHeaderName(config.AuthHeader),
		basculehttp.WithHeaderDelimiter(config.AuthDelimiter),
	)
	handler := alice.New(authConstructor)

	// set up the registerer
	basicConfig := webhookClient.BasicConfig{
		Timeout:         config.WebhookTimeout,
		RegistrationURL: config.WebhookRegistrationURL,
		Request:         config.WebhookRequest,
	}
	registerer, err := webhookClient.NewBasicRegisterer(&acquire.DefaultAcquirer{}, secretGetter, basicConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup registerer: %v\n", err.Error())
		os.Exit(1)
	}
	periodicRegisterer := webhookClient.NewPeriodicRegisterer(registerer, config.WebhookRegistrationInterval, nil)

	// start the registerer
	periodicRegisterer.Start()

	// start listening
	http.Handle(config.Endpoint, handler.ThenFunc(returnStatus(config.ResponseCode)))
	err = http.ListenAndServe(config.Port, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error serving http requests: %v\n", err.Error())
		os.Exit(1)
	}
}

func returnStatus(code int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	}
}
