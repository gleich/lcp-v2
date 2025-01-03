package applemusic

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gleich/lcp-v2/internal/apis"
	"github.com/gleich/lcp-v2/internal/cache"
	"github.com/gleich/lumber/v3"
	"github.com/go-chi/chi/v5"
)

type playlistSummary struct {
	Name       string `json:"name"`
	TrackCount int    `json:"track_count"`
	ID         string `json:"id"`
}

type playlist struct {
	Name         string    `json:"name"`
	Tracks       []song    `json:"tracks"`
	LastModified time.Time `json:"last_modified"`
	ID           string    `json:"id"`
}

type playlistTracksResponse struct {
	Next string         `json:"next"`
	Data []songResponse `json:"data"`
}

type playlistResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			LastModifiedDate time.Time `json:"lastModifiedDate"`
			Name             string    `json:"name"`
		} `json:"attributes"`
	} `json:"data"`
}

func fetchPlaylist(id string) (playlist, error) {
	playlistData, err := sendAppleMusicAPIRequest[playlistResponse](
		fmt.Sprintf("/v1/me/library/playlists/%s", id),
	)
	if err != nil {
		if !errors.Is(err, apis.WarningError) {
			lumber.Error(err, "failed to fetch playlist for", id)
		}
		return playlist{}, err
	}

	var totalResponseData []songResponse
	trackData, err := sendAppleMusicAPIRequest[playlistTracksResponse](
		fmt.Sprintf("/v1/me/library/playlists/%s/tracks", id),
	)
	if err != nil {
		return playlist{}, err
	}
	totalResponseData = append(totalResponseData, trackData.Data...)
	for trackData.Next != "" {
		trackData, err = sendAppleMusicAPIRequest[playlistTracksResponse](trackData.Next)
		if err != nil {
			if !errors.Is(err, apis.WarningError) {
				lumber.Error(err, "failed to paginate through tracks for playlist with id of", id)
			}
			return playlist{}, err
		}
		totalResponseData = append(totalResponseData, trackData.Data...)
	}

	var tracks []song
	for _, t := range totalResponseData {
		tracks = append(tracks, songFromSongResponse(t))
	}

	return playlist{
		Name:         playlistData.Data[0].Attributes.Name,
		LastModified: playlistData.Data[0].Attributes.LastModifiedDate,
		Tracks:       tracks,
		ID:           playlistData.Data[0].ID,
	}, nil
}

func playlistEndpoint(c *cache.Cache[cacheData]) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		c.DataMutex.RLock()
		var p *playlist
		for _, plist := range c.Data.Playlists {
			if plist.ID == id {
				p = &plist
				break
			}
		}

		if p == nil {
			c.DataMutex.RUnlock()
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(p)
		c.DataMutex.RUnlock()
		if err != nil {
			lumber.Error(err, "failed to write json data to request")
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
