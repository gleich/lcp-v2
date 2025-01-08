package steam

import (
	"net/http"
	"time"

	"github.com/gleich/lumber/v3"
	"pkg.mattglei.ch/lcp-2/internal/cache"
)

func Setup(mux *http.ServeMux) {
	games, err := fetchRecentlyPlayedGames()
	if err != nil {
		lumber.Error(err, "initial fetch of games failed")
	}

	steamCache := cache.New("steam", games, err == nil)
	mux.HandleFunc("GET /steam", steamCache.ServeHTTP)
	go steamCache.UpdatePeriodically(fetchRecentlyPlayedGames, 5*time.Minute)
	lumber.Done("setup steam cache")
}
