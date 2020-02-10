package store

import (
	"context"
	"errors"
	"github.com/xmidt-org/webpa-common/logging"
	webhook "github.com/xmidt-org/wrp-listener"
	"sync"
	"time"
)

var (
	errNoHookAvailable = errors.New("no webhook for key")
)

type envelope struct {
	creation time.Time
	hook     webhook.W
}

type InMem struct {
	hooks   map[string]envelope
	lock    sync.RWMutex
	config  InMemConfig
	options *storeConfig
}

func (inMem *InMem) Remove(id string) error {
	// update the store if there is no backend.
	// if it is set. On List() will update the inmem data set
	if inMem.options.backend == nil {
		inMem.lock.Lock()
		delete(inMem.hooks, id)
		inMem.lock.Unlock()
		// update listener
		if inMem.options.listener != nil {
			hooks, _ := inMem.GetWebhook()
			inMem.options.listener.List(hooks)
		}
		return nil
	}
	return inMem.options.backend.Remove(id)
}

func (inMem *InMem) Stop(ctx context.Context) {
	if inMem.options.backend != nil {
		inMem.options.backend.Stop(ctx)
	}
}

func (inMem *InMem) GetWebhook() ([]webhook.W, error) {
	if inMem.options.backend != nil {
		if reader, ok := inMem.options.backend.(Reader); ok {
			return reader.GetWebhook()
		}
	}
	inMem.lock.RLock()
	data := []webhook.W{}
	for _, value := range inMem.hooks {
		if time.Now().Before(value.creation.Add(inMem.config.TTL)) {
			data = append(data, value.hook)
		}
	}
	inMem.lock.RUnlock()
	return data, nil
}

func (inMem *InMem) List(hooks []webhook.W) {
	// update inmem
	if inMem.options.listener != nil {
		inMem.hooks = map[string]envelope{}
		for _, elem := range hooks {
			inMem.hooks[elem.ID()] = envelope{
				creation: time.Now(),
				hook:     elem,
			}
		}
	}
	// TODO: start clean up
	// notify listener
	if inMem.options.listener != nil {
		inMem.options.listener.List(hooks)
	}
}

func (inMem *InMem) Update(w webhook.W) error {
	// update the store if there is no backend.
	// if it is set. On List() will update the inmem data set
	if inMem.options.backend == nil {
		inMem.lock.Lock()
		inMem.hooks[w.ID()] = envelope{
			creation: time.Now(),
			hook:     w,
		}
		inMem.lock.Unlock()
		// update listener
		if inMem.options.listener != nil {
			hooks, _ := inMem.GetWebhook()
			inMem.options.listener.List(hooks)
		}
		return nil
	}
	return inMem.options.backend.Update(w)
}

// CleanUp will free remove old webhooks.
func (inMem *InMem) CleanUp() {
	inMem.lock.Lock()
	for key, value := range inMem.hooks {
		if value.creation.Add(inMem.config.TTL).After(time.Now()) {
			delete(inMem.hooks, key)
		}
	}
	inMem.lock.Unlock()
}

type InMemConfig struct {
	TTL time.Duration
}

// CreateInMemStore will create an inmemory storage that will handle ttl of webhooks.
// listner and back and optional and can be nil
func CreateInMemStore(config InMemConfig, options ...Option) *InMem {
	inMem := &InMem{
		hooks:  map[string]envelope{},
		config: config,
		options: &storeConfig{
			logger: logging.DefaultLogger(),
		},
	}
	for _, o := range options {
		o(inMem.options)
	}
	return inMem
}
