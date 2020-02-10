package dispatch

import (
	"context"
	"github.com/xmidt-org/wrp-go/wrp"
	webhook "github.com/xmidt-org/wrp-listener"
)

type D interface {
	// Dispatch will send the Message to the consumer
	Dispatch(w webhook.W, message wrp.Message) error

	// Stop will stop all threads and cleanup any necessary resources
	Stop(ctx context.Context)
}

type DFunc func(w webhook.W, message wrp.Message) error

func (d DFunc) Dispatch(w webhook.W, message wrp.Message) error {
	return d(w, message)
}

func (d DFunc) Stop(ctx context.Context) {
	// allows for passing a func directly
}
