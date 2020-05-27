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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReasonCode(t *testing.T) {
	tests := []struct {
		description    string
		input          interface{}
		expectedReason ReasonCode
	}{
		{
			description: "Error With Reason",
			input: errWithReason{
				err:    errors.New("test error"),
				reason: DoRequestFail,
			},
			expectedReason: DoRequestFail,
		},
		{
			description:    "Nil Input",
			expectedReason: UnknownReason,
		},
		{
			description:    "Non ReasonCoder",
			input:          3,
			expectedReason: UnknownReason,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			reason := GetReasonCode(tc.input)
			assert.Equal(tc.expectedReason, reason)
		})
	}
}