package store

import (
	"errors"
	webhook "github.com/xmidt-org/wrp-listener"
	"sync"
	"time"
)

var (
	errNoHookAvailable = errors.New("no webhook for key")
)

type envelope struct {
	timestamp time.Time
	hook      webhook.W
}

type InMem struct {
	hooks  map[string]envelope
	lock   sync.RWMutex
	config InMemConfig
}

func (inMem *InMem) GetHooks() []webhook.W {
	inMem.lock.RLock()

	data := []webhook.W{}
	index := 0
	for _, value := range inMem.hooks {
		if value.timestamp.Add(inMem.config.TTL).Before(time.Now()) {
			data[index] = value.hook
		}
	}
	inMem.lock.RUnlock()
	return data
}

func (inMem *InMem) NewList(hooks []webhook.W) error {
	panic("implement me")
}

func (inMem *InMem) Update(w webhook.W) error {

	inMem.lock.Lock()
	inMem.hooks[w.Address] = envelope{
		timestamp: time.Now(),
		hook:      w,
	}
	inMem.lock.Unlock()
	return nil
}

// CleanUp will free remove old webhooks.
func (inMem *InMem) CleanUp() {
	inMem.lock.Lock()
	for key, value := range inMem.hooks {
		if value.timestamp.Add(inMem.config.TTL).After(time.Now()) {
			delete(inMem.hooks, key)
		}
	}
	inMem.lock.Unlock()
}

type InMemConfig struct {
	TTL time.Duration
}

func CreateInMemStore(config InMemConfig) *InMem {
	return &InMem{
		hooks:  map[string]envelope{},
		config: config,
	}
}
