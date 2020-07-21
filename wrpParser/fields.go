package wrpparser

import "github.com/xmidt-org/wrp-go/v2"

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
