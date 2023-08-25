// SPDX-FileCopyrightText: 2020 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package webhookClient

import (
	"errors"
	"time"

	"go.uber.org/zap"
)

// A Registerer attempts to register to a webhook.  If there is a problem, an
// error is returned.
type Registerer interface {
	Register() error
}

// PeriodicRegisterer uses a register to attempt to register at an interval.
// If there is a failure, it will be logged.
type PeriodicRegisterer struct {
	registerer           Registerer
	registrationInterval time.Duration
	logger               *zap.Logger
	measures             *Measures
	shutdown             chan struct{}
}

// NewPeriodicRegisterer creates a registerer that attempts to register at the
// interval given.
func NewPeriodicRegisterer(registerer Registerer, interval time.Duration, logger *zap.Logger, measures *Measures) (*PeriodicRegisterer, error) {
	if interval == 0 {
		return nil, errors.New("interval cannot be 0")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &PeriodicRegisterer{
		registerer:           registerer,
		registrationInterval: interval,
		logger:               logger,
		measures:             measures,
		shutdown:             make(chan struct{}),
	}, nil
}

// Register is just a wrapper to provide the regular Register functionality,
// but generally the periodic registerer should be started and stopped.
func (p *PeriodicRegisterer) Register() error {
	return p.registerer.Register()
}

// Start begins the periodic webhook registration.
func (p *PeriodicRegisterer) Start() {
	go p.registerAtInterval()
}

// Stop stops the periodic webhook registration.
func (p *PeriodicRegisterer) Stop() {
	close(p.shutdown)
}

func (p *PeriodicRegisterer) registerAtInterval() {
	hookagain := time.NewTicker(p.registrationInterval)
	p.registerAndLog()
	for {
		select {
		case <-p.shutdown:
			return
		case <-hookagain.C:
			p.registerAndLog()
		}
	}
}

func (p *PeriodicRegisterer) registerAndLog() {
	err := p.Register()
	if err != nil {
		p.measures.WebhookRegistrationOutcome.With(OutcomeLabel, FailureOutcome, ReasonLabel, GetReasonCode(err).LabelValue()).Add(1.0)
		p.logger.Error("Failed to register webhook", zap.Error(err))
	} else {
		p.measures.WebhookRegistrationOutcome.With(OutcomeLabel, SuccessOutcome, ReasonLabel, "").Add(1.0)
		p.logger.Info("Successfully registered webhook", zap.Error(err))
	}
}
