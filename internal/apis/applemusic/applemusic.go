package applemusic

import (
	"time"

	"github.com/gleich/lcp-v2/internal/cache"
	"github.com/gleich/lumber/v3"
	"github.com/go-chi/chi/v5"
)

const API_ENDPOINT = "https://api.music.apple.com/"

type cacheData struct {
	RecentlyPlayed []song     `json:"recently_played"`
	Playlists      []playlist `json:"playlists"`
}

func cacheUpdate() (cacheData, error) {
	recentlyPlayed, err := fetchRecentlyPlayed()
	if err != nil {
		return cacheData{}, err
	}

	playlistsIDs := []string{
		"p.qQXLxPLtA75zg8e", // 80s
		// "p.LV0PX3EIl0EpDLW", // jazz
		"p.QvDQEebsVbAeokL", // christmas
		"p.AWXoZoxHLrvpJlY", // chill
		"p.LV0PXNoCl0EpDLW", // divorced dad
		// "p.AWXoXPYSLrvpJlY", // alt
		// "p.LV0PXL3Cl0EpDLW", // bops
		// "p.gek1E8efLa68Adp", // classics
		// "p.6xZaArOsvzb5OML", // focus
		// "p.O1kz7EoFVmvz704", // funk
		// "p.V7VYVB0hZo53MQv", // old man
		// "p.QvDQE5RIVbAeokL", // PARTY
		// "p.qQXLxPpFA75zg8e", // RAHHHHHHHH
		// "p.qQXLxpDuA75zg8e", // ROCK
	}
	playlists := []playlist{}
	for _, id := range playlistsIDs {
		playlistData, err := fetchPlaylist(id)
		if err != nil {
			return cacheData{}, err
		}
		playlists = append(playlists, playlistData)
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
	router.Handle("/applemusic/cache/ws", applemusicCache.ServeWS())
	go applemusicCache.StartPeriodicUpdate(cacheUpdate, 30*time.Second)
	lumber.Done("setup apple music cache")
}
