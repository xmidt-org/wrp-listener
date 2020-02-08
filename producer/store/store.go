package store

import webhook "github.com/xmidt-org/wrp-listener"

type Hook interface {
	// Update applies user configurable for registering a webhook
	// i.e. updated the storage with said webhook.
	Update(w webhook.W) error

	// GetHooks return all the current webhooks
	GetHooks()[]webhook.W
}

type Listener interface {
	// NewList will be called when an update is received.
	// aka. new webhook or expired.
	//
	// The list of hooks must contain current webhooks.
	NewList(hooks []webhook.W) error
}
