// SPDX-FileCopyrightText: 2019 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package webhookClient

import (
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// PeriodicRegistererIn is an uber/fx parameter with the in information needed to create a new PeriodicRegisterer.
type PeriodicRegistererIn struct {
	fx.In
	Registerer *BasicRegisterer
	Interval   time.Duration `name:"periodic_registration_interval"`
	Logger     *zap.Logger
	Measures   *Measures
}

// Provide bundles all of the constructors needed to create a new periodic registerer.
func Provide() fx.Option {
	return fx.Options(
		ProvideMetrics(),
		fx.Provide(
			NewBasicRegisterer,
			func(info PeriodicRegistererIn) (*PeriodicRegisterer, error) {
				return NewPeriodicRegisterer(info.Registerer, info.Interval, info.Logger, info.Measures)
			},
		),
	)
}
