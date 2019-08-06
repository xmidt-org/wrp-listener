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
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/spf13/pflag"
	"go.uber.org/fx"

	"github.com/xmidt-org/themis/src/config"
	"github.com/xmidt-org/themis/src/key"
	"github.com/xmidt-org/themis/src/token"
	"github.com/xmidt-org/themis/src/xhttp/xhttpserver"
	"github.com/xmidt-org/themis/src/xlog"
	"github.com/xmidt-org/themis/src/xlog/xloghttp"
	webhook "github.com/xmidt-org/wrp-listener"
)

const (
	applicationName    = "metricListener"
	applicationVersion = "0.0.0"
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

func initialize(e config.Environment) error {
	var (
		file = e.FlagSet.StringP("file", "f", "", "the configuration file to use.  Overrides the search path.")
	)

	e.FlagSet.BoolP("enable-app-logging", "e", false, "enables logging output from the uber/fx App")

	err := e.FlagSet.Parse(e.Arguments)
	if err != nil {
		return err
	}

	switch {
	case len(*file) > 0:
		e.Viper.SetConfigFile(*file)
		err = e.Viper.ReadInConfig()

	default:
		e.Viper.SetConfigName(e.Name)
		e.Viper.AddConfigPath(".")
		e.Viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", e.Name))
		e.Viper.AddConfigPath(fmt.Sprintf("/etc/%s", e.Name))
		err = e.Viper.ReadInConfig()
	}

	if err != nil {
		return err
	}

	return nil
}

func createPrinter(logger log.Logger, e config.Environment) fx.Printer {
	if v, err := e.FlagSet.GetBool("enable-app-logging"); v && err == nil {
		return xlog.Printer{Logger: logger}
	}

	return config.DiscardPrinter{}
}

func main() {
	// // load configuration with viper
	// v := viper.New()
	// v.AddConfigPath(".")
	// v.SetConfigName(applicationName)
	// err := v.ReadInConfig()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to read in viper config: %v\n", err.Error())
	// 	os.Exit(1)
	// }
	// config := new(Config)
	// err = v.Unmarshal(config)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to unmarshal config: %v\n", err.Error())
	// 	os.Exit(1)
	// }

	// // use constant secret for hash
	// secretGetter := secretGetter.NewConstantSecret(config.WebhookRequest.Config.Secret)

	// // set up the middleware
	// htf, err := hashTokenFactory.New("Sha1", sha1.New, secretGetter)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to setup hash token factory: %v\n", err.Error())
	// 	os.Exit(1)
	// }
	// authConstructor := basculehttp.NewConstructor(
	// 	basculehttp.WithTokenFactory("Sha1", htf),
	// 	basculehttp.WithHeaderName(config.AuthHeader),
	// 	basculehttp.WithHeaderDelimiter(config.AuthDelimiter),
	// )
	// handler := alice.New(authConstructor)

	// // set up the registerer
	// basicConfig := webhookClient.BasicConfig{
	// 	Timeout:         config.WebhookTimeout,
	// 	RegistrationURL: config.WebhookRegistrationURL,
	// 	Request:         config.WebhookRequest,
	// }
	// registerer, err := webhookClient.NewBasicRegisterer(&acquire.DefaultAcquirer{}, secretGetter, basicConfig)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to setup registerer: %v\n", err.Error())
	// 	os.Exit(1)
	// }
	// periodicRegisterer := webhookClient.NewPeriodicRegisterer(registerer, config.WebhookRegistrationInterval, nil)

	// // start the registerer
	// periodicRegisterer.Start()

	// // start listening
	// http.Handle(config.Endpoint, handler.ThenFunc(returnStatus(config.ResponseCode)))
	// err = http.ListenAndServe(config.Port, nil)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "error serving http requests: %v\n", err.Error())
	// 	os.Exit(1)
	// }

	var (
		b = config.Bootstrap{
			Name:        applicationName,
			Initializer: initialize,
			Optioners: config.Optioners{
				xlog.Unmarshaller("log", createPrinter),
			},
		}

		app = fx.New(
			b.Provide(),
			provideMetrics("prometheus"),
			fx.Provide(
				token.Unmarshal("token"),
				func() []xloghttp.ParameterBuilder {
					return []xloghttp.ParameterBuilder{
						xloghttp.Method("requestMethod"),
						xloghttp.URI("requestURI"),
						xloghttp.RemoteAddress("remoteAddr"),
					}
				},
				xhttpserver.ProvideParseForm,
				xhttpserver.UnmarshalResponseHeaders("responseHeaders"),
				// provideClient("client"),
			),
			fx.Invoke(
				RunServer("servers.primary"),
				// RunMetricsServer("servers.metrics"),
				xhttpserver.InvokeOptional("servers.pprof", xhttpserver.AddPprofRoutes),
			),
		)
	)

	if err := app.Err(); err != nil {
		if err == pflag.ErrHelp {
			return
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	app.Run()
}

type CommonIn struct {
	fx.In
	ServerMetricsIn

	ParseForm         xhttpserver.ParseForm
	ParameterBuilders []xloghttp.ParameterBuilder `optional:"true"`
}

type KeyServerIn struct {
	xhttpserver.ServerIn
	CommonIn

	Handler key.Handler
}

func RunServer(serverConfigKey string) func(KeyServerIn) error {
	return func(in KeyServerIn) error {
		_, err := xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(ur xhttpserver.UnmarshalResult) error {
				ur.Router.Handle("/key/{kid}", in.Handler).Methods("GET")
				ur.Router.Use(xhttpserver.TrackWriter)
				ur.Router.Use(xloghttp.Logging{Base: ur.Logger, Builders: in.ParameterBuilders}.Then)
				// ur.Router.Use(metricsMiddleware(in.ServerMetricsIn, ur)...)

				return nil
			},
		)

		return err
	}
}

func returnStatus(code int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	}
}
