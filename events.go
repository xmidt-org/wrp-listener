// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

import "time"

// EventRegistration is an event that occurs during webhook registration.
//
// The time the registration is attempted may be quite different from the time
// the event is created.  Therefore it is recorded in the event as At when it
// occurs.
//
// The duration of the registration may be of interest so it is captured in the
// event as Duration when it occurs.
//
// The body of the request or response may be of interest so it is captured in
// the event as Body when an error occurs and it is available.
//
// The status code of the response may be of interest so it is captured in the
// event as StatusCode when it occurs.
//
// Any error that occurs during the registration is captured in the event as Err
// when it occurs.  Multiple error may be included for each event.
type EventRegistration struct {
	// At holds the starting time of the event if applicable.
	At time.Time

	// Duration holds the duration of the event if applicable.
	Duration time.Duration

	// The body of the request or response if applicable.
	Body []byte

	// StatusCode holds the HTTP status code returned by the webhook registration.
	StatusCode int

	// Err holds any error that occurred while performing the registration.
	Err error
}

// RegistrationListener is a sink for registration events.
type RegistrationEventListener interface {
	OnRegistrationEvent(EventRegistration)
}

// TokenizeEvent is an event that occurs during Tokenize() call.
//
// When available the header, algorithms, and algorithm used are included.
// Any error that occurs during tokenization is included.
type TokenizeEvent struct {
	// Header holds the header that was used to tokenize the request.
	Header string

	// Algorithms holds the algorithms that were offered to tokenize the request.
	Algorithms []string

	// Algorithm holds the algorithm that was used to tokenize the request.
	Algorithm string

	// Err holds any error that occurred while tokenizing the request.
	Err error
}

// RegistrationListener is a sink for registration events.
type TokenizeEventListener interface {
	OnTokenizeEvent(TokenizeEvent)
}

// AuthorizeEvent is an event that occurs during the Authorize() call.
// When available the algorithm used is included.
// Any error that occurs during authorization is included.
type AuthorizeEvent struct {
	// Algorithm holds the algorithm that was used for authorization.
	Algorithm string

	// Err holds any error that occurred while tokenizing the request.
	Err error
}

// RegistrationListener is a sink for registration events.
type AuthorizeEventListener interface {
	OnAuthorizeEvent(AuthorizeEvent)
}
