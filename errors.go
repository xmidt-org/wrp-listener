// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"errors"
)

var (
	// ErrInput is returned when an invalid input is provided.
	ErrInput = errors.New("invalid input")

	// ErrInvalidTokenHeader is returned when the token header is invalid.
	ErrInvalidTokenHeader = errors.New("invalid token header")

	// ErrRegistrationFailed is returned when the webhook registration fails.
	ErrRegistrationFailed = errors.New("registration failed")

	// ErrRegistrationNotAttempted is returned when the webhook registration
	// was not attempted.
	ErrRegistrationNotAttempted = errors.New("registration not attempted")

	// ErrNotAcceptedHash is returned when the hash is not accepted, because it
	// is not in the list of accepted hashes.
	ErrNotAcceptedHash = errors.New("not accepted hash")

	// ErrAuthFetchFailed is returned when the auth fetch fails and returns an
	// error.
	ErrAuthFetchFailed = errors.New("auth fetch failed")

	// ErrNewRequestFailed is returned when the request cannot be created.
	ErrNewRequestFailed = errors.New("new request failed")

	// ErrInvalidHeaderFormat is returned when the header is not in the correct
	// format.
	ErrInvalidHeaderFormat = errors.New("invalid header format")

	// ErrInvalidAlgorithm is returned when the algorithm is not supported.
	ErrAlgorithmNotFound = errors.New("algorithm not found")

	// ErrNoToken is returned when the token is not found.
	ErrNoToken = errors.New("no token")

	// ErrInvalidSignature is returned when the signature is invalid.
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrUnableToReadBody is returned when the body cannot be read.
	ErrUnableToReadBody = errors.New("unable to read body")
)
