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
	"context"
	"crypto/sha1"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/goph/emperror"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/xmidt-org/bascule"
	"github.com/xmidt-org/bascule/acquire"
	"github.com/xmidt-org/bascule/basculehttp"
	"github.com/xmidt-org/webpa-common/basculechecks"
	"github.com/xmidt-org/webpa-common/concurrent"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/webpa-common/server"
	webhook "github.com/xmidt-org/wrp-listener"
	"github.com/xmidt-org/wrp-listener/hashTokenFactory"
	secretGetter "github.com/xmidt-org/wrp-listener/secret"
	"github.com/xmidt-org/wrp-listener/webhookClient"
)

const (
	applicationName    = "metricListener"
	applicationVersion = "0.0.0"
	apiBase            = "/api/v1"
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
	JWT                         acquire.JWTAcquirerOptions
}

func SetLogger(logger log.Logger) func(delegate http.Handler) http.Handler {
	return func(delegate http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.WithContext(logging.WithLogger(r.Context(),
					log.With(logger, "requestHeaders", r.Header, "requestURL", r.URL.EscapedPath(), "method", r.Method)))
				delegate.ServeHTTP(w, ctx)
			})
	}
}

func GetLogger(ctx context.Context) bascule.Logger {
	return log.With(logging.GetLogger(ctx), "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
}

func main() {

	start := time.Now()

	var (
		f, v                                   = pflag.NewFlagSet(applicationName, pflag.ContinueOnError), viper.New()
		logger, metricsRegistry, listener, err = server.Initialize(applicationName, os.Args, f, v, basculechecks.Metrics)
		acquirer                               webhookClient.Acquirer
	)

	if parseErr, done := printVersion(f, os.Args); done {
		// if we're done, we're exiting no matter what
		exitIfError(logger, emperror.Wrap(parseErr, "failed to parse arguments"))
		os.Exit(0)
	}

	exitIfError(logger, emperror.Wrap(err, "unable to initialize viper"))
	logging.Info(logger).Log(logging.MessageKey(), "Successfully loaded config file", "configurationFile", v.ConfigFileUsed())

	// load configuration with viper
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

	var m *basculechecks.JWTValidationMeasures

	if metricsRegistry != nil {
		m = basculechecks.NewJWTValidationMeasures(metricsRegistry)
	}
	ml := basculechecks.NewMetricListener(m)

	authConstructor := basculehttp.NewConstructor(
		basculehttp.WithCLogger(GetLogger),
		basculehttp.WithTokenFactory("Sha1", htf),
		basculehttp.WithHeaderName(config.AuthHeader),
		basculehttp.WithHeaderDelimiter(config.AuthDelimiter),
		basculehttp.WithCErrorResponseFunc(ml.OnErrorResponse),
	)
	handler := alice.New(SetLogger(logger), authConstructor, basculehttp.NewListenerDecorator(ml))

	// set up the registerer
	basicConfig := webhookClient.BasicConfig{
		Timeout:         config.WebhookTimeout,
		RegistrationURL: config.WebhookRegistrationURL,
		Request:         config.WebhookRequest,
	}

	acquirer = &acquire.DefaultAcquirer{}

	if config.JWT.AuthURL != "" && config.JWT.Buffer != 0 && config.JWT.Timeout != 0 {
		a := acquire.NewJWTAcquirer(config.JWT)
		acquirer = &a
	}

	registerer, err := webhookClient.NewBasicRegisterer(acquirer, secretGetter, basicConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup registerer: %v\n", err.Error())
		os.Exit(1)
	}
	periodicRegisterer := webhookClient.NewPeriodicRegisterer(registerer, config.WebhookRegistrationInterval, logger)

	// start the registerer
	periodicRegisterer.Start()

	// start listening
	router := mux.NewRouter()
	router.Handle(apiBase+config.Endpoint, handler.ThenFunc(returnStatus(config.ResponseCode)))

	// MARK: Starting the server
	var (
		runnable concurrent.Runnable
		done     <-chan struct{}
		wg       *sync.WaitGroup
		shutdown chan struct{}
	)
	_, runnable, done = listener.Prepare(logger, nil, metricsRegistry, router)
	wg, shutdown, err = concurrent.Execute(runnable)
	exitIfError(logger, emperror.Wrap(err, "unable to start device manager"))

	logging.Info(logger).Log(logging.MessageKey(), fmt.Sprintf("%s is up and running!", applicationName), "elapsedTime", time.Since(start))

	signals := make(chan os.Signal, 10)
	signal.Notify(signals)
	for exit := false; !exit; {
		select {
		case s := <-signals:
			if s != os.Kill && s != os.Interrupt {
				logging.Info(logger).Log(logging.MessageKey(), "ignoring signal", "signal", s)
			} else {
				logging.Error(logger).Log(logging.MessageKey(), "exiting due to signal", "signal", s)
				exit = true
			}
		case <-done:
			logging.Error(logger).Log(logging.MessageKey(), "one or more servers exited")
			exit = true
		}
	}

	periodicRegisterer.Stop()
	time.Sleep(5 * time.Minute)
	close(shutdown)
	wg.Wait()

	logging.Info(logger).Log(logging.MessageKey(), "Listener has shut down")

}

func printVersion(f *pflag.FlagSet, arguments []string) (error, bool) {
	printVer := f.BoolP("version", "v", false, "displays the version number")
	if err := f.Parse(arguments); err != nil {
		return err, true
	}

	if *printVer {
		fmt.Println(applicationVersion)
		return nil, true
	}
	return nil, false
}

func exitIfError(logger log.Logger, err error) {
	if err != nil {
		if logger != nil {
			logging.Error(logger, emperror.Context(err)...).Log(logging.ErrorKey(), err.Error())
		}
		fmt.Fprintf(os.Stderr, "Error: %#v\n", err.Error())
		os.Exit(1)
	}
}

func returnStatus(code int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	}
}
