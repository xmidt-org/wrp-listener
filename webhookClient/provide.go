package webhookClient

import (
	"time"

	"github.com/go-kit/kit/log"
	"go.uber.org/fx"
)

// PeriodicRegistererIn is an uber/fx parameter with the in information needed to create a new PeriodicRegisterer.
type PeriodicRegistererIn struct {
	fx.In
	Registerer *BasicRegisterer
	Interval   time.Duration `name:"periodic_registration_interval"`
	Logger     log.Logger
	Measures   *Measures
}

// BasicRegistererIn is an uber/fx parameter with the information to create a new BasicRegisterer.
type BasicRegistererIn struct {
	fx.In
	Config   BasicConfig
	Acquirer Acquirer
	Secret   SecretGetter
}

// Provide bundles all of the constructors needed to create a new periodic registerer.
func Provide() fx.Option {
	return fx.Options(
		ProvideMetrics(),
		fx.Provide(
			NewProvideMeasures,
			NewBasicRegisterer,
			func(info PeriodicRegistererIn) (*PeriodicRegisterer, error) {
				return NewPeriodicRegisterer(info.Registerer, info.Interval, info.Logger, info.Measures)
			},
		),
	)
}
