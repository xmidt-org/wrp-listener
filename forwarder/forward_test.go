package forwarder

import (
	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/wrp-go/wrp"
	webhook "github.com/xmidt-org/wrp-listener"
	"github.com/xmidt-org/wrp-listener/dispatch"
	"github.com/xmidt-org/wrp-listener/store"
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

	forwader = CreateForwader(func(options ...store.Option) store.Pusher {
		return store.CreateInMemStore(store.InMemConfig{TTL: time.Minute * 5}, options...)
	}, func(w webhook.W) dispatch.D {
		return simpleDispatcher
	}, ForwardMessageToAllWebhooks, logging.NewTestLogger(nil, t))

	_, ok := forwader.(store.Pusher)
	assert.True(ok, "forwarder is not a hook storage")
	_, ok = forwader.(dispatch.D)
	assert.True(ok, "forwarder is not a dispatcher")
	_, ok = forwader.(store.Listener)
	assert.True(ok, "forwarder is not a dispatcher")
	_, ok = forwader.(store.Reader)
	assert.True(ok, "forwarder is not a dispatcher")
}
