package wrpparser

import "github.com/xmidt-org/wrp-go/v2"

type Field int

const (
	Source Field = iota
	Destination
)

func getFieldValue(f Field, msg *wrp.Message) string {
	switch f {
	case Destination:
		return msg.Destination
	default:
		return msg.Source
	}
}
