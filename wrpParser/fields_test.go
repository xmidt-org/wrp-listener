// SPDX-FileCopyrightText: 2020 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/wrp-go/v3"
)

func TestGetFieldValue(t *testing.T) {
	expectedDest := "dest"
	expectedSource := "src"
	testWRP := &wrp.Message{
		Destination: expectedDest,
		Source:      expectedSource,
	}
	dest := getFieldValue(Destination, testWRP)
	src := getFieldValue(Source, testWRP)

	assert := assert.New(t)
	assert.Equal(expectedDest, dest)
	assert.Equal(expectedSource, src)
}
