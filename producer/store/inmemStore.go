package store

import (
	"context"
	"errors"
	"fmt"
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
	hooks     map[string]envelope
	lock      sync.RWMutex
	config    InMemConfig
	listner   Listener
	hookStore Hook
}

func (inMem *InMem) Stop(ctx context.Context) {
	if inMem.hookStore != nil {
		inMem.hookStore.Stop(ctx)
	}
}

func (inMem *InMem) GetHooks() []webhook.W {
	fmt.Println("inMem get hooks before")
	inMem.lock.RLock()

	data := []webhook.W{}
	for _, value := range inMem.hooks {
		fmt.Println(value.timestamp.Add(inMem.config.TTL))
		fmt.Println(time.Now())
		fmt.Println(inMem.config.TTL)
		if time.Now().Before(value.timestamp.Add(inMem.config.TTL)) {
			data = append(data, value.hook)
		}
	}
	inMem.lock.RUnlock()
	fmt.Println("inMem get hooks after", data)
	return data
}

func (inMem *InMem) SetListener(listener Listener) {
	inMem.listner = listener
}

func (inMem *InMem) NewList(hooks []webhook.W) {
	fmt.Println("inmem NewList")
	// update inmem
	if inMem.hookStore != nil {
		inMem.hooks = map[string]envelope{}
		for _, elem := range hooks {
			inMem.hooks[elem.ID()] = envelope{
				timestamp: time.Now(),
				hook:      elem,
			}
		}
	}
	if inMem.listner != nil {
		inMem.listner.NewList(hooks)
	}
}

func (inMem *InMem) Update(w webhook.W) error {
	fmt.Println("inmem update")
	if inMem.hookStore == nil {
		inMem.lock.Lock()
		inMem.hooks[w.ID()] = envelope{
			timestamp: time.Now(),
			hook:      w,
		}
		inMem.lock.Unlock()
		// update listener
		if inMem.listner != nil {
			inMem.listner.NewList(inMem.GetHooks())
		}
		return nil
	}
	return inMem.hookStore.Update(w)
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
