// SPDX-FileCopyrightText: 2020 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package webhookClient

// ReasonCode is a way to describe what went wrong when trying to register for
// a webhook.
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

var unknownReasonLabelValue = "unknown"

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
// or unknown if it's some wacky value.
func (rc ReasonCode) LabelValue() string {
	if lv, ok := reasonLabelValues[rc]; ok {
		return lv
	}

	return unknownReasonLabelValue
}

// ReasonCoder is anything that can return a ReasonCode.
type ReasonCoder interface {
	ReasonCode() ReasonCode
}

// GetReasonCode returns the ReasonCode if the object is a ReasonCoder.
// Otherwise, it returns UnknownReason.
func GetReasonCode(v interface{}) ReasonCode {
	if v == nil {
		return UnknownReason
	}

	if rc, ok := v.(ReasonCoder); ok {
		return rc.ReasonCode()
	}

	return UnknownReason
}
