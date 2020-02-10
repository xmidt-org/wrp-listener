package forwarder

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/xmidt-org/wrp-go/wrp"
	webhook "github.com/xmidt-org/wrp-listener"
	"github.com/xmidt-org/wrp-listener/dispatch"
	"github.com/xmidt-org/wrp-listener/store"
	"sync"
	"sync/atomic"
)

var (
	errStoppedForwader = errors.New("forwarder not accepting anymore messages")
)

type envelope struct {
	hook       webhook.W
	dispatcher dispatch.D
	mark       bool
}

type Forwader struct {
	dispatchers map[string]envelope
	backend     store.Pusher
	forwader    ForwardMessage
	lock        sync.RWMutex
	build       func(w webhook.W) dispatch.D
	stopped     int32
}

type BuilderFunc func(w webhook.W) dispatch.D

func CreateForwader(buildStorage func(...store.Option) store.Pusher, builder func(w webhook.W) dispatch.D, forwader ForwardMessage, logger log.Logger) *Forwader {
	f := &Forwader{
		build:       builder,
		forwader:    forwader,
		stopped:     1,
		dispatchers: map[string]envelope{},
	}
	f.backend = buildStorage(store.WithLogger(logger), store.WithListener(f))
	return f
}

func (f *Forwader) Update(w webhook.W) error {
	return f.backend.Update(w)
}
func (f *Forwader) Remove(id string) error {
	return f.backend.Remove(id)
}
func (f *Forwader) Forward(message wrp.Message) error {
	return f.Dispatch(webhook.W{}, message)
}
func (f *Forwader) Dispatch(_ webhook.W, message wrp.Message) error {
	if atomic.LoadInt32(&f.stopped) == 0 {
		return errStoppedForwader
	}
	f.lock.RLock()
	pushed := false
	for _, envelope := range f.dispatchers {
		if f.forwader(envelope.hook, message) {
			err := envelope.dispatcher.Dispatch(envelope.hook, message)
			fmt.Println(err)
			pushed = true
		}
	}
	f.lock.RUnlock()
	if !pushed {
		return errors.New("no webhooks for event")
	}
	return nil
}

func (f *Forwader) Stop(ctx context.Context) {
	atomic.StoreInt32(&f.stopped, 0)
	f.backend.Stop(ctx)
	for _, hook := range f.dispatchers {
		hook.dispatcher.Stop(ctx)
	}
}

func (f *Forwader) List(hooks []webhook.W) {
	f.lock.Lock()
	for _, hook := range hooks {
		if _, ok := f.dispatchers[hook.ID()]; !ok {
			// if hook does not exist create dispatcher
			dispatcher := f.build(hook)
			if dispatcher != nil {
				f.dispatchers[hook.ID()] = envelope{
					hook:       hook,
					dispatcher: dispatcher,
					mark:       true,
				}
			} else {
				envelope := f.dispatchers[hook.ID()]
				// mark hook
				envelope.mark = true
				// update webhook data
				envelope.hook = hook
			}
		}
	}

	// cleanup old dispatchers
	for key, item := range f.dispatchers {
		if !item.mark {
			go item.dispatcher.Stop(context.Background())
			delete(f.dispatchers, key)
		}
	}
	f.lock.Unlock()
}

func (f *Forwader) GetWebhook() ([]webhook.W, error) {
	if reader, ok := f.backend.(store.Reader); ok {
		return reader.GetWebhook()
	}
	// return current knowledge
	f.lock.RLock()
	data := []webhook.W{}
	for _, value := range f.dispatchers {
		data = append(data, value.hook)
	}
	f.lock.RUnlock()
	return data, nil
}
