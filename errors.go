// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"errors"
)

var (
	ErrInput                    = errors.New("invalid input")
	ErrInvalidAuth              = errors.New("invalid auth")
	ErrInvalidRegistration      = errors.New("invalid registration")
	ErrRegistrationFailed       = errors.New("registration failed")
	ErrRegistrationNotAttempted = errors.New("registration not attempted")
	ErrNotAcceptedHash          = errors.New("not accepted hash")
)
