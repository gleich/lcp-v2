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

	"github.com/gleich/lumber/v3"
	"pkg.mattglei.ch/lcp-v2/internal/apis"
	"pkg.mattglei.ch/lcp-v2/internal/auth"
	"pkg.mattglei.ch/lcp-v2/internal/secrets"
)

type Cache[T any] struct {
	name      string
	DataMutex sync.RWMutex
	Data      T
	Updated   time.Time
	filePath  string
}

func New[T any](name string, data T) *Cache[T] {
	cache := Cache[T]{
		name:     name,
		Updated:  time.Now(),
		filePath: filepath.Join(secrets.SECRETS.CacheFolder, fmt.Sprintf("%s.json", name)),
	}
	cache.loadFromFile()
	cache.Update(data)
	return &cache
}

type CacheResponse[T any] struct {
	Data    T         `json:"data"`
	Updated time.Time `json:"updated"`
}

func (c *Cache[T]) ServeHTTP() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAuthorized(w, r) {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		c.DataMutex.RLock()
		err := json.NewEncoder(w).Encode(CacheResponse[T]{Data: c.Data, Updated: c.Updated})
		c.DataMutex.RUnlock()
		if err != nil {
			err = fmt.Errorf("%v failed to write json data to request", err)
			lumber.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (c *Cache[T]) Update(data T) {
	c.DataMutex.RLock()
	old, err := json.Marshal(c.Data)
	if err != nil {
		lumber.Error(err, "failed to json marshal old data")
		return
	}
	c.DataMutex.RUnlock()
	new, err := json.Marshal(data)
	if err != nil {
		lumber.Error(err, "failed to json marshal new data")
		return
	}

	if string(old) != string(new) && string(new) != "null" && strings.Trim(string(new), " ") != "" {
		c.DataMutex.Lock()
		c.Data = data
		c.Updated = time.Now()
		c.DataMutex.Unlock()

		c.persistToFile()
		lumber.Done(strings.ToUpper(c.name), "cache updated")
	}
}

func (c *Cache[T]) UpdatePeriodically(update func() (T, error), interval time.Duration) {
	for {
		time.Sleep(interval)
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
