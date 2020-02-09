package main

import (
	"fmt"
	"github.com/cenkalti/backoff/v3"
	"github.com/justinas/alice"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/wrp-go/wrp"
	webhook "github.com/xmidt-org/wrp-listener"
	"github.com/xmidt-org/wrp-listener/producer"
	"github.com/xmidt-org/wrp-listener/producer/dispatch"
	"github.com/xmidt-org/wrp-listener/producer/dispatch/queueDispatch"
	"github.com/xmidt-org/wrp-listener/producer/dispatch/retry"
	"github.com/xmidt-org/wrp-listener/producer/forwarder"
	"github.com/xmidt-org/wrp-listener/producer/store"
	"net/http"
	"os"
	"time"
)

func main() {
	handler := alice.New()

	logger := logging.DefaultLogger()
	logging.Info(logger).Log(logging.MessageKey(), "hmm")

	storage := store.CreateInMemStore(store.InMemConfig{TTL: time.Minute * 5})

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
	app := producer.WebhookServer{Forwader: forwarder.CreateForwader(storage, func(_ webhook.W) dispatch.D {
		fmt.Println("builder")
		return dispatcher
	}, func(w webhook.W, message wrp.Message) bool {
		fmt.Println("forwader")
		return true
	})}
	http.Handle("/hook", handler.ThenFunc(app.RegisterListener))
	http.Handle("/hooks", handler.ThenFunc(app.GetSanatizedWebhooks))
	go stream(app.Forwader)
	err := http.ListenAndServe(":7100", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error serving http requests: %v\n", err.Error())
		os.Exit(1)
	}
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
