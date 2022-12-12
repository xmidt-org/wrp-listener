package main

import "github.com/xmidt-org/bascule/acquire"

// determineTokenAcquirer always returns a valid TokenAcquirer
func determineTokenAcquirer(config WebhookConfig) (acquire.Acquirer, error) {
	defaultAcquirer := &acquire.DefaultAcquirer{}
	if config.JWT.AuthURL != "" && config.JWT.Buffer != 0 && config.JWT.Timeout != 0 {
		return acquire.NewRemoteBearerTokenAcquirer(config.JWT)
	}

	if config.Basic != "" {
		return acquire.NewFixedAuthAcquirer("Basic " + config.Basic)
	}

	return defaultAcquirer, nil
}
