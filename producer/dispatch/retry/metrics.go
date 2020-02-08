package retry

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/xmidt-org/webpa-common/xmetrics"
)

const (
	RetryCounter = "retry_dispatch_count"
	EndCounter   = "retry_dispatch_end_counter"
)

// Metrics returns the Metrics relevant to this package
func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		{
			Name:       RetryCounter,
			Type:       "counter",
			Help:       "The total number of dispatch messages retried",
			LabelNames: []string{},
		},
		{
			Name:       EndCounter,
			Type:       "counter",
			Help:       "the total number of dispatched messages that are done, no more retrying",
			LabelNames: []string{},
		},
	}
}

type Measures struct {
	RetryCounter metrics.Counter
	EndCounter   metrics.Counter
}

func NewMeasures(p provider.Provider) Measures {
	return Measures{
		RetryCounter: p.NewCounter(RetryCounter),
		EndCounter:   p.NewCounter(EndCounter),
	}
}
