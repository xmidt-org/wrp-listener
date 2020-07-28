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

package wrpparser

import (
	"github.com/stretchr/testify/mock"
	"github.com/xmidt-org/wrp-go/v3"
)

type MockDeviceFinder struct {
	mock.Mock
}

func (f *MockDeviceFinder) FindDeviceID(msg *wrp.Message) (string, error) {
	args := f.Called(msg)
	return args.String(0), args.Error(1)
}

type MockClassifier struct {
	mock.Mock
}

func (c *MockClassifier) Label(msg *wrp.Message) (string, bool) {
	args := c.Called(msg)
	return args.String(0), args.Bool(1)
}
