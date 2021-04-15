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
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/touchstone/touchkit"
	"github.com/xmidt-org/webpa-common/xmetrics"
	"go.uber.org/fx"
)

//Names for our metrics
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

//Metrics returns the Metrics relevant to this package
func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		xmetrics.Metric{
			Name:       WebhookRegistrationOutcome,
			Type:       xmetrics.CounterType,
			Help:       "Counter for the periodic registerer, providing the outcome of a registration attempt",
			LabelNames: []string{OutcomeLabel, ReasonLabel},
		},
	}
}

//Measures describes the defined metrics that will be used by clients.
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

// NewProvideMeasures converts MeasuresIn to Measures
func NewProvideMeasures(in MeasuresIn) *Measures {
	return &Measures{
		WebhookRegistrationOutcome: in.WebhookRegistrationOutcome,
	}
}

// ProvideMetrics provides the metrics relevant to this package as uber/fx options.
func ProvideMetrics() fx.Option {
	return touchkit.Counter(
		prometheus.CounterOpts{
			Name: WebhookRegistrationOutcome,
			Help: "Counter for the periodic registerer, providing the outcome of a registration attempt",
		}, OutcomeLabel, ReasonLabel,
	)
}
