// SPDX-FileCopyrightText: 2020 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package webhookClient

// implements error and ReasonCoder
type errWithReason struct {
	err    error
	reason ReasonCode
}

func (e errWithReason) Error() string {
	return e.err.Error()
}

func (e errWithReason) ReasonCode() ReasonCode {
	return e.reason
}
