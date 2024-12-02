package applemusic

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gleich/lcp-v2/internal/apis"
	"github.com/gleich/lumber/v3"
)

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
		fmt.Sprintf("v1/me/library/playlists/%s", id),
	)
	if err != nil {
		lumber.Error(err, "failed to fetch playlist for", id)
		return playlist{}, err
	}

	var totalResponseData []songResponse
	trackData, err := sendAppleMusicAPIRequest[playlistTracksResponse](
		fmt.Sprintf("v1/me/library/playlists/%s/tracks", id),
	)
	if err != nil {
		lumber.Error(err, "failed to get tracks for playlist with id of", id)
		return playlist{}, err
	}
	totalResponseData = append(totalResponseData, trackData.Data...)
	for trackData.Next != "" {
		req, err := http.NewRequest("GET", trackData.Next, nil)
		if err != nil {
			lumber.Error(err, "failed to create pagination request")
			return playlist{}, err
		}
		trackData, err = apis.SendRequest[playlistTracksResponse](req)
		if err != nil {
			lumber.Error(err, "failed to paginate through tracks for playlist with id of", id)
			return playlist{}, err
		}
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
