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

import (
	"encoding/json"
	"errors"
	"net"
	"time"
)

// W is the structure that represents the Webhook listener
// data we share.
//
// (Note to Wes: this follows the golang naming conventions.  webhook.Webhook "stutters",
// and this type is really the central type of this package.  Calling it a single letter is the norm.
// This could also go in the server package, in which case I'd change the name to Webhook, since
// service.Webhook works better.  See https://blog.golang.org/package-names)
type W struct {
	// Configuration for message delivery
	Config Config `json:"config"`

	// The URL to notify when we cut off a client due to overflow.
	// Optional, set to "" to disable behavior
	FailureURL string `json:"failure_ur,omitempty"`

	// The list of regular expressions to match event type against.
	Events []string `json:"events,omitempty"`

	// Matcher type contains values to match against the metadata.
	Matcher Matcher `json:"matcher,omitempty"`

	// The specified duration for this hook to live
	Duration time.Duration `json:"duration,omitempty"`

	// The absolute time when this hook is to be disabled
	Until time.Time `json:"until,omitempty"`

	// The address that performed the registration
	Address string `json:"registered_from_address,omitempty"`
}

// Configuration for message delivery
type Config struct {
	// The URL to deliver messages to.
	URL string `json:"url"`

	// The content-type to set the messages to (unless specified by WRP).
	ContentType string `json:"content_type,omitempty"`

	// The secret to use for the SHA1 HMAC.
	// Optional, set to "" to disable behavior.
	Secret string `json:"secret,omitempty"`

	// The max number of times to retry for webhook
	MaxRetryCount int `json:"max_retry_count,omitempty"`

	// alt_urls is a list of explicit URLs that should be round robin on faliure
	AlternativeURLs []string `json:"alt_urls,omitempty"`
}

// Matcher type contains values to match against the metadata.
type Matcher struct {
	// The list of regular expressions to match device id type against.
	DeviceID []string `json:"device_id"`
}

func NewW(jsonString []byte, ip string) (w *W, err error) {
	w = new(W)

	err = json.Unmarshal(jsonString, w)
	if err != nil {
		var wa []W

		err = json.Unmarshal(jsonString, &wa)
		if err != nil {
			return
		}
		w = &wa[0]
	}

	err = w.sanitize(ip)
	if nil != err {
		w = nil
	}
	return
}

func (w *W) sanitize(ip string) (err error) {

	if "" == w.Config.URL {
		err = errors.New("invalid Config URL")
		return
	}

	if 0 == len(w.Events) {
		err = errors.New("invalid events")
		return
	}

	// TODO Validate content type ?  What about different types?

	if 0 == len(w.Matcher.DeviceID) {
		w.Matcher.DeviceID = []string{".*"} // match anything
	}

	if "" == w.Address && "" != ip {
		// Record the IP address the request came from
		host, _, _err := net.SplitHostPort(ip)
		if nil != _err {
			err = _err
			return
		}
		w.Address = host
	}

	// always set duration to default
	w.Duration = time.Minute * 5

	if &w.Until == nil || w.Until.Equal(time.Time{}) {
		w.Until = time.Now().Add(w.Duration)
	}

	return
}

// ID creates the canonical string identifing a WebhookListener
func (w *W) ID() string {
	return w.Config.URL
}
