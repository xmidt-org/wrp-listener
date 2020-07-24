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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewStrParser(t *testing.T) {
	mockFinder := new(MockDeviceFinder)
	mockClassifier := new(MockClassifier)
	options := []ParserOption{
		WithDeviceFinder("label", mockFinder),
		WithDeviceFinder("label", nil),
	}
	tests := []struct {
		description       string
		classifier        Classifier
		finder            DeviceFinder
		expectedStrParser *StrParser
		expectedErr       error
	}{
		{
			description: "Success",
			classifier:  mockClassifier,
			finder:      mockFinder,
			expectedStrParser: &StrParser{
				classifier:    mockClassifier,
				defaultFinder: mockFinder,
				finders:       map[string]DeviceFinder{"label": mockFinder},
			},
		},
		{
			description: "Nil Finder Error",
			classifier:  mockClassifier,
			expectedErr: errNilFinder,
		},
		{
			description: "Nil Classifier Error",
			finder:      mockFinder,
			expectedErr: errNilClassifier,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			p, err := NewStrParser(tc.classifier, tc.finder, options...)
			assert.Equal(tc.expectedStrParser, p)
			assert.Equal(tc.expectedErr, err)
		})
	}
}

func TestParse(t *testing.T) {
	testErr := errors.New("test finder error")
	labels := []string{"label1", "label2", "label3"}
	goodID := "meh"
	type classifierResp struct {
		label string
		ok    bool
	}
	tests := []struct {
		description               string
		deviceFinderErr           bool
		classifierResponse        classifierResp
		defaultDeviceFinderCalled bool
		expectedErr               error
	}{
		{
			description: "Success",
			classifierResponse: classifierResp{
				label: "label2",
				ok:    true,
			},
		},
		{
			description: "Success Without Label",
			classifierResponse: classifierResp{
				label: "",
				ok:    false,
			},
			defaultDeviceFinderCalled: true,
		},
		{
			description: "Success Without Finder For Label",
			classifierResponse: classifierResp{
				label: "nonexistent label",
				ok:    true,
			},
			defaultDeviceFinderCalled: true,
		},
		{
			description:               "Finder Error",
			deviceFinderErr:           true,
			defaultDeviceFinderCalled: true,
			expectedErr:               testErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			// set up mocks
			mockClassifier := new(MockClassifier)
			mockClassifier.On("Label", mock.Anything).Return(tc.classifierResponse.label, tc.classifierResponse.ok).Once()
			mockFinderCalled := new(MockDeviceFinder)
			if tc.deviceFinderErr {
				mockFinderCalled.On("FindDeviceID", mock.Anything).Return("", testErr)
			} else {
				mockFinderCalled.On("FindDeviceID", mock.Anything).Return(goodID, nil)
			}
			mockFinderNotCalled := new(MockDeviceFinder)

			// set up string parser
			p := &StrParser{
				classifier:    mockClassifier,
				defaultFinder: mockFinderNotCalled,
				finders:       map[string]DeviceFinder{},
			}
			// set our device finders
			if tc.defaultDeviceFinderCalled {
				p.defaultFinder = mockFinderCalled
			}
			for _, l := range labels {
				if l == tc.classifierResponse.label {
					p.finders[l] = mockFinderCalled
				} else {
					p.finders[l] = mockFinderNotCalled
				}
			}

			id, err := p.Parse(nil)
			mockClassifier.AssertExpectations(t)
			mockFinderCalled.AssertExpectations(t)
			mockFinderNotCalled.AssertExpectations(t)
			if tc.deviceFinderErr {
				assert.Empty(id)
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Equal(goodID, id)
				assert.Nil(err)
			}

		})
	}
}
