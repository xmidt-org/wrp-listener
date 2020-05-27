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

package webhookClient

// ReasonCode is easier to consume programmatically if it's an enumerated integral type
type ReasonCode int

const (
	UnknownReason ReasonCode = iota
	GetSecretFail
	MarshalRequestFail
	AcquireJWTFail
	CreateRequestFail
	DoRequestFail
	ReadBodyFail
	Non200Response
)

// the catch-all value, intern'ed as a variable
var unknownReasonLabelValue = "unknown"

// using a global map essentially interns the string values
// it also makes it easy to inspect the code when adding/removing reasons
var reasonLabelValues = map[ReasonCode]string{
	GetSecretFail:      "get_secret_failed",
	MarshalRequestFail: "marshal_request_failed",
	AcquireJWTFail:     "acquire_jwt_failed",
	CreateRequestFail:  "create_request_failed",
	DoRequestFail:      "do_request_failed",
	ReadBodyFail:       "read_body_failed",
	Non200Response:     "non_200_response",
}

// LabelValue returns the metric label value for this reason code,
// or unknown if it's some wacky value
func (rc ReasonCode) LabelValue() string {
	if lv, ok := reasonLabelValues[rc]; ok {
		return lv
	}

	return unknownReasonLabelValue
}

// Anything, not just errors, can implement this interface.  Similar to the pattern in packages
// like go-kit, e.g. StatusCoder
type ReasonCoder interface {
	ReasonCode() ReasonCode
}

// You can then provide a helper function that accepts any old object, including errors:
func GetReasonCode(v interface{}) ReasonCode {
	if rc, ok := v.(ReasonCoder); ok {
		return rc.ReasonCode()
	}

	return UnknownReason
}
