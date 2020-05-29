/**
 * Copyright 2020 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package webhookClient

import (
	"time"

	"github.com/xmidt-org/webpa-common/logging"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics/provider"
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
	logger               log.Logger
	measures             *Measures
	shutdown             chan struct{}
}

var (
	defaultLogger = log.NewNopLogger()
)

// NewPeriodicRegisterer creates a registerer that attempts to register at the
// interval given.
func NewPeriodicRegisterer(registerer Registerer, interval time.Duration, logger log.Logger, provider provider.Provider) *PeriodicRegisterer {
	if logger == nil {
		logger = defaultLogger
	}

	m := NewMeasures(provider)

	return &PeriodicRegisterer{
		registerer:           registerer,
		registrationInterval: interval,
		logger:               logger,
		measures:             m,
	}
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
		p.logger.Log(level.Key(), level.ErrorValue(), logging.MessageKey(), "Failed to register webhook",
			logging.ErrorKey(), err.Error())
	} else {
		p.measures.WebhookRegistrationOutcome.With(OutcomeLabel, SuccessOutcome, ReasonLabel, "").Add(1.0)
		p.logger.Log(level.Key(), level.InfoValue(), logging.MessageKey(), "Successfully registered webhook")
	}
}
