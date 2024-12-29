package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gleich/lcp-v2/internal/apis"
	"github.com/gleich/lcp-v2/internal/metrics"
	"github.com/gleich/lcp-v2/internal/secrets"
	"github.com/gleich/lumber/v3"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Cache[T any] struct {
	name            string
	dataMutex       sync.RWMutex
	data            T
	updated         time.Time
	updateCounter   prometheus.Counter
	requestCounter  prometheus.Counter
	filePath        string
	wsConnPool      map[*websocket.Conn]bool
	wsConnPoolMutex sync.Mutex
	wsUpgrader      websocket.Upgrader
}

func New[T any](name string, data T) *Cache[T] {
	cache := Cache[T]{
		name: name,
		updateCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("cache_%s_updates", name),
			Help: fmt.Sprintf(`The total number of times the cache "%s" has been updated`, name),
		}),
		requestCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("cache_%s_requests", name),
			Help: fmt.Sprintf(`The total number of times the cache "%s" has been requested`, name),
		}),
		filePath:   filepath.Join(secrets.SECRETS.CacheFolder, fmt.Sprintf("%s.json", name)),
		wsConnPool: make(map[*websocket.Conn]bool),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	cache.loadFromFile()
	cache.Update(data)
	return &cache
}

type cacheData[T any] struct {
	Data    T         `json:"data"`
	Updated time.Time `json:"updated"`
}

// Handle a GET request to load data from the given cache
func (c *Cache[T]) ServeHTTP() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.dataMutex.RLock()
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(cacheData[T]{Data: c.data, Updated: c.updated})
		c.dataMutex.RUnlock()
		c.requestCounter.Inc()
		if err != nil {
			lumber.Error(err, "failed to write data")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

// Update the given cache
func (c *Cache[T]) Update(data T) {
	var updated bool
	c.dataMutex.Lock()
	old, err := json.Marshal(c.data)
	if err != nil {
		lumber.Error(err, "failed to json marshal old data")
		return
	}
	new, err := json.Marshal(data)
	if err != nil {
		lumber.Error(err, "failed to json marshal new data")
		return
	}
	if string(old) != string(new) && string(new) != "null" && strings.Trim(string(new), " ") != "" {
		c.data = data
		c.updated = time.Now()
		updated = true
	}
	c.dataMutex.Unlock()

	if updated {
		c.updateCounter.Inc()
		metrics.CacheUpdates.Inc()
		c.persistToFile()
		connectionsUpdated := c.broadcastUpdate()
		if connectionsUpdated == 0 {
			lumber.Done(strings.ToUpper(c.name), "cache updated")
		} else {
			lumber.Done(
				strings.ToUpper(c.name),
				"cache updated;",
				"updated", connectionsUpdated, "websocket connections",
			)
		}
	}
}

func (c *Cache[T]) UpdatePeriodically(updateFunc func() (T, error), interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		data, err := updateFunc()
		if err != nil {
			if !errors.Is(err, apis.WarningError) {
				lumber.Error(err, "updating", c.name, "cache failed")
			}
		} else {
			c.Update(data)
		}
	}
}
