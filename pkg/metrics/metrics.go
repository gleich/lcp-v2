package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CacheUpdates = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_updates",
		Help: "The total number of cache updates",
	})
)
