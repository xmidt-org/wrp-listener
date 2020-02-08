package dispatch

import (
	"github.com/xmidt-org/wrp-go/wrp"
	webhook "github.com/xmidt-org/wrp-listener"
)

type D interface {
	// Dispatch will send the Message to the consumer
	Dispatch(w webhook.W, message wrp.Message) error
}

type DFunc func(w webhook.W, message wrp.Message) error

func (d DFunc) Dispatch(w webhook.W, message wrp.Message) error {
	return d(w, message)
}
