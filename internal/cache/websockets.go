package cache

import (
	"net/http"

	"github.com/gleich/lumber/v3"
	"github.com/gorilla/websocket"
)

// Handle websocket connections
func (c *Cache[T]) ServeWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := c.wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			lumber.Error(err, "failed to upgrade connection to websocket")
			return
		}
		c.wsConnPoolMutex.Lock()
		c.wsConnPool[conn] = true
		c.wsConnPoolMutex.Unlock()

		// sending initial data
		c.dataMutex.RLock()
		err = conn.WriteJSON(c.data)
		c.dataMutex.RUnlock()
		if err != nil {
			lumber.Error(err, "failed to write initial cache data for", c.name)
			c.removeConnection(conn)
			return
		}

		// spawning goroutine to handle connection
		go func() {
			defer c.removeConnection(conn)
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
			}
		}()
	}
}

func (c *Cache[T]) broadcastUpdate() int {
	c.dataMutex.RLock()
	d := c.data
	c.dataMutex.RUnlock()

	updatedConnections := 0
	for conn := range c.wsConnPool {
		err := conn.WriteJSON(d)
		if err != nil {
			lumber.Error(err, "failed to broadcast update to client")
			c.removeConnection(conn)
		} else {
			updatedConnections++
		}
	}
	return updatedConnections
}

func (c *Cache[T]) removeConnection(conn *websocket.Conn) {
	c.wsConnPoolMutex.Lock()
	delete(c.wsConnPool, conn)
	c.wsConnPoolMutex.Unlock()
	conn.Close()
}
