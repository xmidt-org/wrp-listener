// SPDX-FileCopyrightText: 2020 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package webhookClient

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/touchstone/touchkit"

	// nolint:staticcheck
	"github.com/xmidt-org/webpa-common/v2/xmetrics"
	"go.uber.org/fx"
)

// Names for our metrics
const (
	WebhookRegistrationOutcome = "webhook_registration"
)

// labels
const (
	OutcomeLabel = "outcome"
	ReasonLabel  = "reason"
)

const (
	// outcomes
	SuccessOutcome = "success"
	FailureOutcome = "failure"
)

// Metrics returns the Metrics relevant to this package
func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		{
			Name:       WebhookRegistrationOutcome,
			Type:       xmetrics.CounterType,
			Help:       "Counter for the periodic registerer, providing the outcome of a registration attempt",
			LabelNames: []string{OutcomeLabel, ReasonLabel},
		},
	}
}

// Measures describes the defined metrics that will be used by clients.
type Measures struct {
	WebhookRegistrationOutcome metrics.Counter
}

// MeasuresIn is an uber/fx parameter with the webhook registration counter
type MeasuresIn struct {
	fx.In
	WebhookRegistrationOutcome metrics.Counter `name:"webhook_registration"`
}

// NewMeasures realizes desired metrics.
func NewMeasures(p provider.Provider) *Measures {
	return &Measures{
		WebhookRegistrationOutcome: p.NewCounter(WebhookRegistrationOutcome),
	}
}

// ProvideMetrics provides the metrics relevant to this package as uber/fx options.
func ProvideMetrics() fx.Option {
	return fx.Options(
		touchkit.Counter(
			prometheus.CounterOpts{
				Name: WebhookRegistrationOutcome,
				Help: "Counter for the periodic registerer, providing the outcome of a registration attempt",
			}, OutcomeLabel, ReasonLabel,
		),
		fx.Provide(
			func(in MeasuresIn) *Measures {
				return &Measures{
					WebhookRegistrationOutcome: in.WebhookRegistrationOutcome,
				}
			},
		),
	)
}
