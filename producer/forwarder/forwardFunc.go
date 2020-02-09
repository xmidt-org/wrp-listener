package forwarder

import (
	"github.com/xmidt-org/wrp-go/wrp"
	webhook "github.com/xmidt-org/wrp-listener"
)

// ForwardMessage func allows for ability to set custom logic for if a message should be forwarded to the webhook.
type ForwardMessage func(w webhook.W, message wrp.Message) bool

func ForwardMessageToAllWebhooks(_ webhook.W, _ wrp.Message) bool {
	return true
}
