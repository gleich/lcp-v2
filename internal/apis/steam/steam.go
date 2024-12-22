package steam

import (
	"time"

	"github.com/gleich/lcp-v2/internal/cache"
	"github.com/gleich/lumber/v3"
	"github.com/go-chi/chi/v5"
)

func Setup(router *chi.Mux) {
	games, err := fetchRecentlyPlayedGames()
	if err != nil {
		lumber.Fatal(err, "initial fetch of games failed")
	}

	steamCache := cache.NewCache("steam", games)
	router.Get("/steam/cache", steamCache.ServeHTTP())
	router.Handle("/steam/cache/ws", steamCache.ServeWS())
	go steamCache.StartPeriodicUpdate(fetchRecentlyPlayedGames, 5*time.Minute)
	lumber.Done("setup steam cache")
}
