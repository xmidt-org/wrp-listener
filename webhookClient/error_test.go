// SPDX-FileCopyrightText: 2020 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package webhookClient

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrWithReason(t *testing.T) {
	var testObj interface{} = errWithReason{
		err:    errors.New("test error with reason"),
		reason: CreateRequestFail,
	}
	_, ok := testObj.(ReasonCoder)
	assert.True(t, ok)

	_, ok = testObj.(error)
	assert.True(t, ok)
}
