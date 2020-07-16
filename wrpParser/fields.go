package wrpparser

import "github.com/xmidt-org/wrp-go/v2"

type field int

const (
	Source field = iota
	Destination
)

func getFieldValue(f field, msg *wrp.Message) string {
	switch f {
	case Destination:
		return msg.Destination
	default:
		return msg.Source
	}
}
