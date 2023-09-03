// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"fmt"
	"strings"
	"time"
)

// Registration is an event that occurs during webhook registration.
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
type Registration struct {
	// At holds the starting time of the event if applicable.
	At time.Time

	// Duration holds the duration of the event if applicable.
	Duration time.Duration

	// The body of the request or response if applicable.
	Body []byte

	// StatusCode holds the HTTP status code returned by the webhook registration.
	StatusCode int

	// Until holds the time the registration expires if applicable.
	Until time.Time

	// Err holds any error that occurred while performing the registration.
	Err error
}

func (r Registration) String() string {
	buf := strings.Builder{}

	buf.WriteString("event.Registration{\n")
	buf.WriteString(fmt.Sprintf("  At:         %s\n", r.At.Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("  Duration:   %s\n", r.Duration.String()))
	buf.WriteString(fmt.Sprintf("  Body:       '%s'\n", string(r.Body)))
	buf.WriteString(fmt.Sprintf("  StatusCode: %d\n", r.StatusCode))
	buf.WriteString(fmt.Sprintf("  Until:      %s\n", r.Until.Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("  Err:        %v\n", r.Err))
	buf.WriteString("}\n")

	return buf.String()
}

// RegistrationListener is a sink for registration events.
type RegistrationListener interface {
	OnRegistrationEvent(Registration)
}

// RegistrationFunc is a function that implements the RegistrationListener
// interface.  It is useful for creating a listener from a function.
type RegistrationFunc func(Registration)

func (f RegistrationFunc) OnRegistrationEvent(r Registration) {
	f(r)
}

// Tokenize is an event that occurs during Tokenize() call.
//
// When available the header, algorithms, and algorithm used are included.
// Any error that occurs during tokenization is included.
type Tokenize struct {
	// Header holds the header that was used to tokenize the request.
	Header string

	// Algorithms holds the algorithms that were offered to tokenize the request.
	Algorithms []string

	// Algorithm holds the algorithm that was used to tokenize the request.
	Algorithm string

	// Err holds any error that occurred while tokenizing the request.
	Err error
}

func (t Tokenize) String() string {
	buf := strings.Builder{}

	buf.WriteString("event.Tokenize{\n")
	buf.WriteString(fmt.Sprintf("  Header:     '%s'\n", t.Header))
	buf.WriteString(fmt.Sprintf("  Algorithms: [%s]\n", strings.Join(t.Algorithms, ", ")))
	buf.WriteString(fmt.Sprintf("  Algorithm:  '%s'\n", t.Algorithm))
	buf.WriteString(fmt.Sprintf("  Err:        %v\n", t.Err))
	buf.WriteString("}\n")

	return buf.String()
}

// RegistrationListener is a sink for registration events.
type TokenizeListener interface {
	OnTokenizeEvent(Tokenize)
}

// TokenizeEventFunc is a function that implements the TokenizeListener
// interface.  It is useful for creating a listener from a function.
type TokenizeFunc func(Tokenize)

func (f TokenizeFunc) OnTokenizeEvent(t Tokenize) {
	f(t)
}

// AuthorizeEvent is an event that occurs during the Authorize() call.
// When available the algorithm used is included.
// Any error that occurs during authorization is included.
type Authorize struct {
	// Algorithm holds the algorithm that was used for authorization.
	Algorithm string

	// Err holds any error that occurred while tokenizing the request.
	Err error
}

func (a Authorize) String() string {
	buf := strings.Builder{}

	buf.WriteString("event.Authorize{\n")
	buf.WriteString(fmt.Sprintf("  Algorithm:  '%s'\n", a.Algorithm))
	buf.WriteString(fmt.Sprintf("  Err:        %v\n", a.Err))
	buf.WriteString("}\n")

	return buf.String()
}

// RegistrationListener is a sink for registration events.
type AuthorizeListener interface {
	OnAuthorizeEvent(Authorize)
}

// AuthorizeFunc is a function that implements the AuthorizeListener
// interface.  It is useful for creating a listener from a function.
type AuthorizeFunc func(Authorize)

func (f AuthorizeFunc) OnAuthorizeEvent(a Authorize) {
	f(a)
}
