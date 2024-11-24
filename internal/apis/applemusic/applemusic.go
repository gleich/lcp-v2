package applemusic

import (
	"time"

	"github.com/gleich/lcp-v2/internal/cache"
	"github.com/gleich/lumber/v3"
	"github.com/go-chi/chi/v5"
)

type cacheData struct {
	RecentlyPlayed []song              `json:"recently_played"`
	Playlists      map[string]playlist `json:"playlists"`
}

func cacheUpdate() (cacheData, error) {
	recentlyPlayed, err := fetchRecentlyPlayed()
	if err != nil {
		return cacheData{}, err
	}

	playlistsIDs := []string{
		"p.AWXoZoxHLrvpJlY", // CHILL
		"p.qQXLxPLtA75zg8e", // HIGHSCHOOL 1989
		"p.LV0PXNoCl0EpDLW", // DIVORCED DAD
	}
	playlists := map[string]playlist{}
	for _, id := range playlistsIDs {
		playlistData, err := fetchPlaylist(id)
		if err != nil {
			return cacheData{}, err
		}
		playlists[id] = playlistData
	}

	return cacheData{
		RecentlyPlayed: recentlyPlayed,
		Playlists:      playlists,
	}, nil
}

func Setup(router *chi.Mux) {
	data, err := cacheUpdate()
	if err != nil {
		lumber.Fatal(err, "initial fetch of cache data failed")
	}

	applemusicCache := cache.NewCache("applemusic", data)
	router.Get("/applemusic/cache", applemusicCache.ServeHTTP())
	go applemusicCache.StartPeriodicUpdate(cacheUpdate, 1*time.Minute)
	lumber.Done("setup apple music cache")
}
