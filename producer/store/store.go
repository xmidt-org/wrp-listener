package store

import (
	"context"
	webhook "github.com/xmidt-org/wrp-listener"
)

type Hook interface {
	// Update applies user configurable for registering a webhook
	// i.e. updated the storage with said webhook.
	Update(w webhook.W) error

	// GetHooks return all the current webhooks
	GetHooks() []webhook.W

	// Stop will stop all threads and cleanup any necessary resources
	Stop(context context.Context)

	// SetListener will update the internal listener.
	SetListener(listener Listener)
}

type Listener interface {
	// NewList will be called when an update is received.
	// aka. new webhook or expired.
	//
	// The list of hooks must contain current webhooks.
	NewList(hooks []webhook.W)
}

type ListnerFunc func(hooks []webhook.W)

func (listner ListnerFunc) NewList(hooks []webhook.W) {
	listner(hooks)
}
