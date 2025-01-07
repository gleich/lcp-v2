package applemusic

import (
	"encoding/json"
	"net/http"
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
		"p.gek1E8efLa68Adp", // classics
		"p.LV0PX3EIl0EpDLW", // jazz
		"p.AWXoZoxHLrvpJlY", // chill
		"p.V7VYVB0hZo53MQv", // old man
		"p.qQXLxPLtA75zg8e", // 80s
		"p.LV0PXNoCl0EpDLW", // divorced dad
		"p.AWXoXPYSLrvpJlY", // alt
		"p.QvDQE5RIVbAeokL", // PARTY
		"p.LV0PXL3Cl0EpDLW", // bops
		"p.6xZaArOsvzb5OML", // focus
		"p.O1kz7EoFVmvz704", // funk
		"p.qQXLxPpFA75zg8e", // RAHHHHHHHH
		"p.qQXLxpDuA75zg8e", // ROCK
		"p.O1kz7zbsVmvz704", // country
		"p.QvDQEN0IVbAeokL", // fall
		// "p.ZOAXAMZF4KMD6ob", // sad girl music
		// "p.QvDQEebsVbAeokL", // christmas
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

	applemusicCache := cache.New("applemusic", data)
	router.Get("/applemusic", serveHTTP(applemusicCache))
	router.Get("/applemusic/playlists/{id}", playlistEndpoint(applemusicCache))
	go applemusicCache.UpdatePeriodically(cacheUpdate, 30*time.Second)
	lumber.Done("setup apple music cache")
}

type cacheDataResponse struct {
	PlaylistSummaries []playlistSummary `json:"playlist_summaries"`
	RecentlyPlayed    []song            `json:"recently_played"`
}

func serveHTTP(c *cache.Cache[cacheData]) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		c.DataMutex.RLock()

		data := cacheDataResponse{}
		for _, p := range c.Data.Playlists {
			firstFourTracks := []song{}
			for _, track := range p.Tracks {
				if len(firstFourTracks) < 4 {
					firstFourTracks = append(firstFourTracks, track)
				}
			}
			data.PlaylistSummaries = append(
				data.PlaylistSummaries,
				playlistSummary{
					Name:            p.Name,
					ID:              p.ID,
					TrackCount:      len(p.Tracks),
					FirstFourTracks: firstFourTracks,
				},
			)
		}
		data.RecentlyPlayed = c.Data.RecentlyPlayed

		err := json.NewEncoder(w).
			Encode(cache.CacheResponse[cacheDataResponse]{Data: data, Updated: c.Updated})
		c.DataMutex.RUnlock()
		if err != nil {
			lumber.Error(err, "failed to write json data to request")
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
