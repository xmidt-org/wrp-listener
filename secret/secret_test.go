// SPDX-FileCopyrightText: 2019 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package secretGetter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstantSecret(t *testing.T) {
	assert := assert.New(t)
	expectedSecret := "test secret"
	cs := NewConstantSecret(expectedSecret)
	secret, err := cs.GetSecret()
	assert.Nil(err)
	assert.Equal(expectedSecret, secret)
}
