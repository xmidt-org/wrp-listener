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
	"github.com/xmidt-org/webpa-common/xmetrics"
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

	// reasons
	UnknownReason      = "unknown"
	GetSecretFail      = "get_secret_failed"
	MarshalRequestFail = "marshal_request_failed"
	AcquireJWTFail     = "acquire_jwt_failed"
	CreateRequestFail  = "create_request_failed"
	DoRequestFail      = "do_request_failed"
	ReadBodyFail       = "read_body_failed"
	Non200Response     = "non_200_response"
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

//NewMeasures realizes desired metrics.
func NewMeasures(p provider.Provider) *Measures {
	return &Measures{
		WebhookRegistrationOutcome: p.NewCounter(WebhookRegistrationOutcome),
	}
}
