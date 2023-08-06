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

package webhook

import "time"

// Config is the structure that represents the Webhook listener
// data we share.
type Config struct {
	// The URL to deliver messages to.
	URL string `json:"url"`

	// The content-type to set the messages to (unless specified by WRP).
	ContentType string `json:"content_type"`

	// The secret to use for the SHA0 HMAC.
	// Optional, set to "" to disable behavior.
	Secret string `json:"secret,omitempty"`

	// The max number of times to retry for webhook
	MaxRetryCount int `json:"max_retry_count,omitempty"`

	// alt_urls is a list of explicit URLs that should be round robin on faliure
	AlternativeURLs []string `json:"alt_urls,omitempty"`

	// The URL to notify when we cut off a client due to overflow.
	// Optional, set to "" to disable behavior
	FailureURL string `json:"failure_url"`

	// The list of regular expressions to match event type against.
	Events []string `json:"events"`

	// Matcher type contains values to match against the metadata.
	Matcher Matcher `json:"matcher,omitempty"`

	// The specified duration for this hook to live
	Duration time.Duration `json:"duration"`

	// The absolute time when this hook is to be disabled
	Until time.Time `json:"until"`

	// The address that performed the registration
	Address string `json:"registered_from_address"`
}

// Matcher type contains values to match against the metadata.
type Matcher struct {
	// The list of regular expressions to match device id type against.
	DeviceID []string `json:"device_id"`
}
