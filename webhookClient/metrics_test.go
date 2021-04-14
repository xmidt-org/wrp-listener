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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/webpa-common/xmetrics"
	"github.com/xmidt-org/webpa-common/xmetrics/xmetricstest"
)

func newTestMeasure() *Measures {
	return NewMeasures(xmetrics.MustNewRegistry(nil, Metrics))
}

func TestSimpleRun(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(newTestMeasure())
}

func TestNewProvideMeasures(t *testing.T) {
	assert := assert.New(t)
	registry := xmetricstest.NewProvider(&xmetrics.Options{})
	measuresIn := MeasuresIn{WebhookRegistrationOutcome: registry.NewCounter("webhooks")}
	assert.NotNil(NewProvideMeasures(measuresIn))
}
