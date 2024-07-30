package steam

import (
	"time"

	"github.com/gleich/lcp-v2/internal/cache"
	"github.com/gleich/lumber/v2"
	"github.com/go-chi/chi/v5"
)

func Setup(router *chi.Mux) {
	games := fetchRecentlyPlayedGames()
	steamCache := cache.NewCache("steam", games)
	router.Get("/steam/cache", steamCache.ServeHTTP())
	go steamCache.StartPeriodicUpdate(fetchRecentlyPlayedGames, 10*time.Minute)
	lumber.Success("setup steam cache")
}
