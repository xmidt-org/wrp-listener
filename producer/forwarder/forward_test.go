package forwarder

import (
	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/wrp-go/wrp"
	webhook "github.com/xmidt-org/wrp-listener"
	"github.com/xmidt-org/wrp-listener/producer/dispatch"
	"github.com/xmidt-org/wrp-listener/producer/store"
	"testing"
	"time"
)

func TestForwarderInterface(t *testing.T) {
	assert := assert.New(t)
	var (
		forwader         interface{}
		simpleDispatcher dispatch.DFunc
	)
	simpleDispatcher = func(w webhook.W, message wrp.Message) error {
		t.Log(message.Destination)
		return nil
	}

	forwader = CreateForwader(store.CreateInMemStore(store.InMemConfig{TTL: time.Second}), func(w webhook.W) dispatch.D {
		return simpleDispatcher
	}, ForwardMessageToAllWebhooks)

	_, ok := forwader.(store.Hook)
	assert.True(ok, "forwarder is not a hook storage")
	_, ok = forwader.(dispatch.D)
	assert.True(ok, "forwarder is not a dispatcher")
	_, ok = forwader.(store.Listener)
	assert.True(ok, "forwarder is not a dispatcher")
}
