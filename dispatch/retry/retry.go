package retry

import (
	"context"
	"github.com/cenkalti/backoff/v3"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/xmidt-org/wrp-go/wrp"
	webhook "github.com/xmidt-org/wrp-listener"
	"github.com/xmidt-org/wrp-listener/dispatch"
	"time"
)

type retryConfig struct {
	backoffConfig backoff.ExponentialBackOff
	measures      *Measures
}

// Option is the function used to configure the retry object.
type Option func(r *retryConfig)

// WithBackoff sets the exponential backoff to use when retrying.  If this
// isn't called, we use the backoff package's default ExponentialBackoff
// configuration.  If any values are considered invalid, they are replaced with
// those defaults.
func WithBackoff(b backoff.ExponentialBackOff) Option {
	return func(r *retryConfig) {
		r.backoffConfig = b
		if r.backoffConfig.InitialInterval < 0 {
			r.backoffConfig.InitialInterval = backoff.DefaultInitialInterval
		}
		if r.backoffConfig.RandomizationFactor < 0 {
			r.backoffConfig.RandomizationFactor = backoff.DefaultRandomizationFactor
		}
		if r.backoffConfig.Multiplier < 1 {
			r.backoffConfig.Multiplier = backoff.DefaultMultiplier
		}
		if r.backoffConfig.MaxInterval < 0 {
			r.backoffConfig.MaxInterval = backoff.DefaultMaxInterval
		}
		if r.backoffConfig.MaxElapsedTime < 0 {
			r.backoffConfig.MaxElapsedTime = backoff.DefaultMaxElapsedTime
		}
		if r.backoffConfig.Clock == nil {
			r.backoffConfig.Clock = backoff.SystemClock
		}
	}
}

// WithMeasures sets a provider to use for metrics.
func WithMeasures(p provider.Provider) Option {
	return func(r *retryConfig) {
		if p != nil {
			m := NewMeasures(p)
			r.measures = &m
		}
	}
}

type RetryDispatcher struct {
	config     retryConfig
	dispatcher dispatch.D
}

func (r *RetryDispatcher) Stop(ctx context.Context) {
	panic("implement me")
}

// AddRetryMetric is a function to add to our metrics when we retry.  The
// function is passed to the backoff package and is called when we are retrying.
func (r *RetryDispatcher) AddRetryMetric(_ error, _ time.Duration) {
	if r.config.measures != nil {
		r.config.measures.RetryCounter.Add(1.0)
	}
}

// Dispatch uses the internal dispatcher to send messages and uses the
// ExponentialBackoff to try again if inserting fails.
func (r *RetryDispatcher) Dispatch(w webhook.W, message wrp.Message) error {

	dispatchFunc := func() error {
		return r.dispatcher.Dispatch(w, message)
	}

	b := r.config.backoffConfig

	err := backoff.RetryNotify(dispatchFunc, &b, r.AddRetryMetric)
	if r.config.measures != nil {
		r.config.measures.EndCounter.Add(1.0)
	}
	return err
}

// CreateRetryInsertService takes an inserter and the options provided and
// creates a RetryInsertService.
func CreateRetryDispatcher(dispatcher dispatch.D, options ...Option) dispatch.D {
	ris := &RetryDispatcher{
		dispatcher: dispatcher,
		config: retryConfig{
			backoffConfig: *backoff.NewExponentialBackOff(),
		},
	}
	for _, o := range options {
		o(&ris.config)
	}
	return ris
}
