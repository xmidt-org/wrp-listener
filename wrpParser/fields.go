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

import "github.com/xmidt-org/wrp-go/v3"

// Field is an enum that describes a specific field in a wrp message.
// Further docs on wrps can be found here:
// https://pkg.go.dev/github.com/xmidt-org/wrp-go@v1.3.4/wrp?tab=doc#Message
type Field int

const (
	// Source is a wrp message's Source field.
	Source Field = iota
	// Destination is a wrp message's Destination field.
	Destination
)

// getFieldValue takes a field and a message, returning the value at the field
// in the message.
func getFieldValue(f Field, msg *wrp.Message) string {
	switch f {
	case Destination:
		return msg.Destination
	default:
		return msg.Source
	}
}
