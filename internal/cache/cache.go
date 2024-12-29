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
	"github.com/gleich/lcp-v2/internal/secrets"
	"github.com/gleich/lumber/v3"
	"github.com/gorilla/websocket"
)

type Cache[T any] struct {
	name            string
	dataMutex       sync.RWMutex
	data            T
	updated         time.Time
	filePath        string
	wsConnPool      map[*websocket.Conn]bool
	wsConnPoolMutex sync.Mutex
	wsUpgrader      websocket.Upgrader
}

func New[T any](name string, data T) *Cache[T] {
	cache := Cache[T]{
		name:       name,
		updated:    time.Now(),
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

func (c *Cache[T]) ServeHTTP() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		c.dataMutex.RLock()
		err := json.NewEncoder(w).Encode(cacheData[T]{Data: c.data, Updated: c.updated})
		c.dataMutex.RUnlock()
		if err != nil {
			lumber.Error(err, "failed to write data")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func (c *Cache[T]) Update(data T) {
	c.dataMutex.RLock()
	old, err := json.Marshal(c.data)
	if err != nil {
		lumber.Error(err, "failed to json marshal old data")
		return
	}
	c.dataMutex.RUnlock()
	new, err := json.Marshal(data)
	if err != nil {
		lumber.Error(err, "failed to json marshal new data")
		return
	}

	if string(old) != string(new) && string(new) != "null" && strings.Trim(string(new), " ") != "" {
		c.dataMutex.Lock()
		c.data = data
		c.updated = time.Now()
		c.dataMutex.Unlock()

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

func (c *Cache[T]) UpdatePeriodically(update func() (T, error), interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		data, err := update()
		if err != nil {
			if !errors.Is(err, apis.WarningError) {
				lumber.Error(err, "updating", c.name, "cache failed")
			}
		} else {
			c.Update(data)
		}
	}
}
