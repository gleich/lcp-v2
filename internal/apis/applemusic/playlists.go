package applemusic

import (
	"path"
	"time"

	"github.com/gleich/lumber/v3"
)

type playlist struct {
	Name         string    `json:"name"`
	Tracks       []song    `json:"tracks"`
	LastModified time.Time `json:"last_modified"`
}

type playlistTracksResponse struct {
	Next string         `json:"next"`
	Data []songResponse `json:"data"`
}

type playlistResponse struct {
	Data []struct {
		Attributes struct {
			LastModifiedDate time.Time `json:"lastModifiedDate"`
			Name             string    `json:"name"`
		} `json:"attributes"`
	} `json:"data"`
}

func fetchPlaylist(id string) (playlist, error) {
	playlistData, err := sendAPIRequest[playlistResponse](path.Join("v1/me/library/playlists/", id))
	if err != nil {
		lumber.Error(err, "failed to fetch playlist for", id)
		return playlist{}, err
	}

	var totalResponseData []songResponse
	trackData, err := sendAPIRequest[playlistTracksResponse](
		path.Join("v1/me/library/playlists/", id, "tracks"),
	)
	if err != nil {
		lumber.Error(err, "failed to get tracks for playlist with id of", id)
	}
	totalResponseData = append(totalResponseData, trackData.Data...)
	for trackData.Next != "" {
		trackData, err = sendAPIRequest[playlistTracksResponse](trackData.Next)
		if err != nil {
			lumber.Error(err, "failed to paginate through tracks for playlist with id of", id)
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
	}, nil
}
