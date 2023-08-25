// SPDX-FileCopyrightText: 2019 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package webhookClient

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockAcquirer struct {
	mock.Mock
}

func (a *MockAcquirer) Acquire() (string, error) {
	args := a.Called()
	return args.String(0), args.Error(1)
}

type MockRoundTripper struct {
	mock.Mock
}

func (rt *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := rt.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

type mockSecretGetter struct {
	mock.Mock
}

func (sg *mockSecretGetter) GetSecret() (string, error) {
	args := sg.Called()
	return args.String(0), args.Error(1)
}
