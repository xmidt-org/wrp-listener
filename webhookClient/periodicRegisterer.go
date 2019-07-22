/**
 * Copyright 2019 Comcast Cable Communications Management, LLC
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

	"github.com/goph/emperror"
	"github.com/xmidt-org/webpa-common/logging"

	"github.com/go-kit/kit/log"
)

type Registerer interface {
	Register() error
}

type PeriodicRegisterer struct {
	registerer           Registerer
	registrationInterval time.Duration
	logger               log.Logger
	shutdown             chan struct{}
}

var (
	defaultLogger = log.NewNopLogger()
)

func NewPeriodicRegisterer(registerer Registerer, interval time.Duration, logger log.Logger) *PeriodicRegisterer {
	if logger == nil {
		logger = defaultLogger
	}
	return &PeriodicRegisterer{
		registerer:           registerer,
		registrationInterval: interval,
		logger:               logger,
	}
}

func (p *PeriodicRegisterer) Register() error {
	return p.registerer.Register()
}

func (p *PeriodicRegisterer) Start() {
	go p.registerAtInterval()
}

func (p *PeriodicRegisterer) Stop() {
	close(p.shutdown)
}

func (p *PeriodicRegisterer) registerAtInterval() {
	hookagain := time.NewTicker(p.registrationInterval)
	err := p.Register()
	if err != nil {
		logging.Error(p.logger, emperror.Context(err)...).Log(logging.MessageKey(), "Failed to register webhook",
			logging.ErrorKey(), err.Error())
	} else {
		logging.Info(p.logger).Log(logging.MessageKey(), "Successfully registered webhook")
	}
	for {
		select {
		case <-p.shutdown:
			return
		case <-hookagain.C:
			err := p.Register()
			if err != nil {
				logging.Error(p.logger, emperror.Context(err)...).Log(logging.MessageKey(), "Failed to register webhook",
					logging.ErrorKey(), err.Error())
			} else {
				logging.Info(p.logger).Log(logging.MessageKey(), "Successfully registered webhook")
			}
		}
	}
}
