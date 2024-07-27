package cache

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/gleich/lcp-v2/pkg/metrics"
	"github.com/gleich/lcp-v2/pkg/secrets"
	"github.com/gleich/lumber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Cache[T any] struct {
	Name           string
	mutex          sync.RWMutex
	data           T
	updated        time.Time
	updateCounter  prometheus.Counter
	requestCounter prometheus.Counter
	filePath       string
}

func New[T any](name string, data T) *Cache[T] {
	cache := Cache[T]{
		Name: name,
		updateCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("cache_%s_updates", name),
			Help: fmt.Sprintf(`The total number of times the cache "%s" has been updated`, name),
		}),
		requestCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("cache_%s_requests", name),
			Help: fmt.Sprintf(`The total number of times the cache "%s" has been requested`, name),
		}),
		filePath: filepath.Join(cacheFolder, fmt.Sprintf("%s.json", name)),
	}
	cache.seedFromFile()
	cache.Update(data)
	return &cache
}

type cacheData[T any] struct {
	Data    T         `json:"data"`
	Updated time.Time `json:"updated"`
}

// Handle a GET request to load data from the given cache
func (c *Cache[T]) Route() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+secrets.SECRETS.ValidToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		c.mutex.RLock()
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(cacheData[T]{Data: c.data, Updated: c.updated})
		c.mutex.RUnlock()
		c.requestCounter.Inc()
		if err != nil {
			lumber.Error(err, "failed to write data")
		}
	})
}

// Update the given cache
func (c *Cache[T]) Update(data T) {
	var updated bool
	c.mutex.Lock()
	if !reflect.DeepEqual(data, c.data) {
		c.data = data
		c.updated = time.Now()
		updated = true
	}
	c.mutex.Unlock()
	c.updateCounter.Inc()
	metrics.CacheUpdates.Inc()
	c.writeToFile()
	if updated {
		lumber.Success(c.Name, "updated")
	}
}

func (c *Cache[T]) PeriodicUpdate(updateFunc func() T, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		c.Update(updateFunc())
	}
}
