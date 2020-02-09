package dispatch

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/wrp-go/wrp"
	webhook "github.com/xmidt-org/wrp-listener"
)

type logDispatcher struct {
	dispatcher D
	logger     log.Logger
}

func (l *logDispatcher) Dispatch(w webhook.W, message wrp.Message) error {
	l.logger.Log(logging.MessageKey(), "dispatching message", "message", message, "webhook", w)
	if l.dispatcher != nil {
		return l.dispatcher.Dispatch(w, message)
	}
	return nil
}

func (l *logDispatcher) Stop(ctx context.Context) {
	if l.dispatcher != nil {
		l.dispatcher.Stop(ctx)
	}
}

func CreateLogDispatcher(logger log.Logger, dispatcher D) D {
	return &logDispatcher{
		dispatcher: dispatcher,
		logger:     logger,
	}
}
