package steam

import (
	"time"

	"github.com/gleich/lumber/v3"
	"github.com/go-chi/chi/v5"
	"pkg.mattglei.ch/lcp-v2/internal/cache"
)

func Setup(router *chi.Mux) {
	games, err := fetchRecentlyPlayedGames()
	if err != nil {
		lumber.Fatal(err, "initial fetch of games failed")
	}

	steamCache := cache.New("steam", games)
	router.Get("/steam", steamCache.ServeHTTP())
	go steamCache.UpdatePeriodically(fetchRecentlyPlayedGames, 5*time.Minute)
	lumber.Done("setup steam cache")
}
