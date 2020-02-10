package queueDispatch

import (
	"context"
	"errors"
	"github.com/xmidt-org/webpa-common/semaphore"
	"github.com/xmidt-org/wrp-go/wrp"
	"github.com/xmidt-org/wrp-listener"
	"github.com/xmidt-org/wrp-listener/dispatch"
	"sync"
)

var (
	errFullQueue = errors.New("queue is full")
)

type QueueDispatcher struct {
	dispatcher dispatch.D
	config     QueueDispatchConfig
	stop       chan struct{}
	wg         sync.WaitGroup
	workers    semaphore.Interface
	jobs       chan func()
}

func (q *QueueDispatcher) Dispatch(w webhook.W, message wrp.Message) error {
	select {
	// TODO:// should we add a timeout to add to the channel?
	case q.jobs <- func() {
		// TODO:// do something with error case
		q.dispatcher.Dispatch(w, message)
	}:
		// TODO:// should this request block? and return the answer
		return nil
	default:
		return errFullQueue
	}
}

func (q *QueueDispatcher) work() {
	defer q.wg.Done()

}

type QueueDispatchConfig struct {
	MaxWorkers int
	QueueSize  int
}

// Stop closes the internal queue and waits for the workers to finish
// processing what has already been added.  This can block as it waits for
// everything to stop.  After Stop() is called, Insert() should not be called
// again, or there will be a panic.
// TODO: ensure consumers can't cause a panic?
func (queue *QueueDispatcher) Stop(ctx context.Context) {
	close(queue.jobs)
	queue.wg.Wait()

	// Grab all the workers to make sure they are done.
	for i := 0; i < queue.config.MaxWorkers; i++ {
		queue.workers.Acquire()
	}
	queue.dispatcher.Stop(ctx)
}

func CreateQueueDispatcher(config QueueDispatchConfig, dispatcher dispatch.D) dispatch.D {

	queue := &QueueDispatcher{
		dispatcher: dispatcher,
		config:     config,
	}

	// start workers
	for i := 0; i < queue.config.MaxWorkers; i++ {
		queue.wg.Add(1)
		go queue.work()
	}

	return dispatcher
}
