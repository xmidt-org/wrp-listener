package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/metrics/provider"
	"github.com/justinas/alice"
	"github.com/spf13/viper"
	"github.com/xmidt-org/bascule/acquire"
	"github.com/xmidt-org/bascule/basculehttp"
	"github.com/xmidt-org/interpreter"
	"github.com/xmidt-org/wrp-go/v3"
	webhook "github.com/xmidt-org/wrp-listener"
	"github.com/xmidt-org/wrp-listener/hashTokenFactory"
	secretGetter "github.com/xmidt-org/wrp-listener/secret"
	"github.com/xmidt-org/wrp-listener/webhookClient"
	"go.uber.org/zap"
)

const (
	applicationName, apiBase = "timedListener", "/api/v1"
)

// WebhookConfig for creating the webhook registration
type WebhookConfig struct {
	RegistrationInterval time.Duration
	Timeout              time.Duration
	RegistrationURL      string
	HostToRegister       string
	Request              webhook.W
	JWT                  acquire.RemoteBearerTokenAcquirerOptions
	Basic                string
}

// SecretConfig for validating the incoming request
type SecretConfig struct {
	Header    string
	Delimiter string
}

// ServerConfig for configuring the listen server
type ServerConfig struct {
	Address string
}

// Config is the central location for the timedListenerConfig
type Config struct {
	Webhook      WebhookConfig
	Secret       SecretConfig
	Server       ServerConfig
	TimeToListen time.Duration
}

// timedListener is only active for the configured time to live
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

	// build json logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err.Error())
		os.Exit(1)
	}

	// use constant secret for hash
	secretGetter := secretGetter.NewConstantSecret(config.Webhook.Request.Config.Secret)

	// set up the middleware
	htf, err := hashTokenFactory.New("sha1", sha1.New, secretGetter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup hash token factory: %v\n", err.Error())
		os.Exit(1)
	}
	authConstructor := basculehttp.NewConstructor(
		basculehttp.WithTokenFactory("sha1", htf),
		basculehttp.WithHeaderName(config.Secret.Header),
		basculehttp.WithHeaderDelimiter(config.Secret.Delimiter),
	)
	handler := alice.New(authConstructor)

	// set up the registerer
	basicConfig := webhookClient.BasicConfig{
		Timeout:         config.Webhook.Timeout,
		RegistrationURL: config.Webhook.RegistrationURL,
		Request:         config.Webhook.Request,
	}

	// get acquirer
	acquirer, err := determineTokenAcquirer(config.Webhook)
	if err != nil {
		logger.Error("failed to determine token acquirer", zap.Error(err))
	}

	registerer, err := webhookClient.NewBasicRegisterer(acquirer, secretGetter, basicConfig)
	if err != nil {
		logger.Error("failed to setup registerer", zap.Error(err))
		os.Exit(1)
	}
	periodicRegisterer, err := webhookClient.NewPeriodicRegisterer(registerer, config.Webhook.RegistrationInterval, logger, webhookClient.NewMeasures(provider.NewDiscardProvider()))
	if err != nil {
		logger.Error("failed to periodic registerer", zap.Error(err))
		os.Exit(1)
	}
	// start the registerer
	periodicRegisterer.Start()

	// start listening
	http.Handle(apiBase+"/events", handler.ThenFunc(handleEventWithLogger(logger)))
	go func() {
		err = http.ListenAndServe(config.Server.Address, nil)
		logger.Error("listener stopped", zap.Error(err))

	}()

	// wait for TimeToListen before last registration call
	waitDuration := config.Webhook.RegistrationInterval + config.Webhook.Request.Duration
	time.Sleep(config.TimeToListen - waitDuration)

	// stop registerer
	periodicRegisterer.Stop()
	time.Sleep(waitDuration)

	// end of program
}

func handleEventWithLogger(logger *zap.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg wrp.Message
		var err error
		msgBytes, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			logger.Error("failed to read body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = wrp.NewDecoderBytes(msgBytes, wrp.Msgpack).Decode(&msg)
		if err != nil {
			logger.Error("failed to decode body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// create interpreted event
		event, err := interpreter.NewEvent(msg)
		if err != nil {
			logger.Error("failed to create interpreted event", zap.Error(err))
			w.WriteHeader(http.StatusOK)
			return
		}

		// print out deviceID
		deviceID, err := event.DeviceID()
		if err != nil {
			logger.Error("failed to get device id", zap.Error(err))
			w.WriteHeader(http.StatusOK)
			return
		}
		// print out eventType
		eventType, err := event.EventType()
		if err != nil {
			logger.Error("failed to get event type", zap.Error(err))
			w.WriteHeader(http.StatusOK)
			return
		}
		fmt.Printf("deviceID: %s, eventType:  %s\n", deviceID, eventType)

		w.WriteHeader(http.StatusOK)
	}
}
