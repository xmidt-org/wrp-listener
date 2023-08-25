// SPDX-FileCopyrightText: 2020 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
