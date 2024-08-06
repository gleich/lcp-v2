package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/gleich/lumber/v2"
)

const cacheFolder = "/Users/matt/Desktop/caches/"

func (c *Cache[T]) persistToFile() {
	var file *os.File
	if _, err := os.Stat(c.filePath); os.IsNotExist(err) {
		folder := filepath.Dir(c.filePath)
		err := os.MkdirAll(folder, 0700)
		if err != nil {
			lumber.Error(err, "failed to create folder at path:", folder)
			return
		}
		file, err = os.Create(c.filePath)
		if err != nil {
			lumber.Error(err, "failed to create file at path:", c.filePath)
			return
		}
	} else {
		file, err = os.OpenFile(c.filePath, os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			lumber.Error(err, "failed to read file at path:", c.filePath)
			return
		}
	}
	defer file.Close()

	b, err := json.Marshal(cacheData[T]{
		Data:    c.data,
		Updated: c.updated,
	})
	if err != nil {
		lumber.Error(err, "encoding data to json failed")
		return
	}
	_, err = file.Write(b)
	if err != nil {
		lumber.Error(err, "writing data to json failed")
	}
}

// Load the cache from the persistent cache file
// returns if the cache was able to be loaded or not
func (c *Cache[T]) loadFromFile() {
	if _, err := os.Stat(c.filePath); !os.IsNotExist(err) {
		b, err := os.ReadFile(c.filePath)
		if err != nil {
			lumber.Error(err, "reading from cache file from", c.filePath, "failed")
			return
		}

		var data cacheData[T]
		err = json.Unmarshal(b, &data)
		if err != nil {
			lumber.Error(err, "unmarshal json data failed from:", string(b))
			return
		}

		c.data = data.Data
		c.updated = data.Updated
	}
}
