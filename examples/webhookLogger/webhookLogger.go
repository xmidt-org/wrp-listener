package main

import (
	"encoding/json"
	"fmt"
	"github.com/cenkalti/backoff/v3"
	"github.com/justinas/alice"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/wrp-go/wrp"
	webhook "github.com/xmidt-org/wrp-listener"
	"github.com/xmidt-org/wrp-listener/dispatch"
	"github.com/xmidt-org/wrp-listener/dispatch/queueDispatch"
	"github.com/xmidt-org/wrp-listener/dispatch/retry"
	"github.com/xmidt-org/wrp-listener/forwarder"
	"github.com/xmidt-org/wrp-listener/store"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func main() {
	handler := alice.New()

	logger := logging.DefaultLogger()
	retryDispatcher := retry.CreateRetryDispatcher(dispatch.CreateLogDispatcher(logging.Info(logger), nil),
		retry.WithBackoff(backoff.ExponentialBackOff{
			InitialInterval:     0,
			RandomizationFactor: .1,
			Multiplier:          1,
			MaxInterval:         1,
			MaxElapsedTime:      10,
		}))
	dispatcher := queueDispatch.CreateQueueDispatcher(queueDispatch.QueueDispatchConfig{
		MaxWorkers: 1,
		QueueSize:  10,
	}, retryDispatcher)
	forwader := forwarder.CreateForwader(
		func(options ...store.Option) store.Pusher {
			return store.CreateInMemStore(store.InMemConfig{TTL: time.Minute * 5}, options...)
		},
		func(_ webhook.W) dispatch.D {
			return dispatcher
		},
		func(w webhook.W, message wrp.Message) bool {
			return true
		},
		logger,
	)

	app := App{Forwader: forwader}
	http.Handle("/hook", handler.ThenFunc(app.HandleRegister))
	http.Handle("/hooks", handler.ThenFunc(app.HandleGetHooks))
	go stream(app.Forwader)
	err := http.ListenAndServe(":7100", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error serving http requests: %v\n", err.Error())
		os.Exit(1)
	}
}

type App struct {
	Forwader *forwarder.Forwader
}

func (app *App) HandleRegister(responseWriter http.ResponseWriter, request *http.Request) {
	payload, err := ioutil.ReadAll(request.Body)
	request.Body.Close()

	w, err := webhook.NewW(payload, request.RemoteAddr)
	if err != nil {
		panic(err)
		responseWriter.Header().Add("X-Wrp-Listener-Error", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = app.Forwader.Update(*w)
	if err != nil {
		panic(err)

		responseWriter.Header().Add("X-Wrp-Listener-Error", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.WriteHeader(http.StatusOK)
}
func (app *App) HandleGetHooks(responseWriter http.ResponseWriter, request *http.Request) {
	hooks, err := app.Forwader.GetWebhook()
	if err != nil {
		responseWriter.Header().Add("X-Wrp-Listener-Error", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(&hooks)
	if err != nil {
		responseWriter.Header().Add("X-Wrp-Listener-Error", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: sanatize input
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.Write(data)
	responseWriter.WriteHeader(http.StatusOK)
}

func stream(dispatcher *forwarder.Forwader) {
	for {
		err := dispatcher.Forward(wrp.Message{
			Source:      "me",
			Destination: time.Now().String(),
			ContentType: "see",
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		time.Sleep(time.Second)
	}
}
