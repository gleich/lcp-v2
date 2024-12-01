package applemusic

import (
	"net/http"
	"net/url"

	"github.com/gleich/lcp-v2/internal/apis"
	"github.com/gleich/lumber/v3"
)

type recentlyPlayedResponse struct {
	Data []songResponse `json:"data"`
}

func fetchRecentlyPlayed() ([]song, error) {
	u, err := url.JoinPath(API_ENDPOINT, "v1/me/recent/played/tracks")
	if err != nil {
		lumber.Error(err, "failed to create URl")
		return []song{}, err
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		lumber.Error(err, "failed to create request")
		return []song{}, err
	}
	response, err := apis.SendRequest[recentlyPlayedResponse](req)
	if err != nil {
		return []song{}, err
	}

	var songs []song
	for _, s := range response.Data {
		songs = append(songs, songFromSongResponse(s))
	}

	// filter out duplicate songs
	seen := make(map[string]bool)
	uniqueSongs := []song{}
	for _, song := range songs {
		if !seen[song.ID] {
			seen[song.ID] = true
			uniqueSongs = append(uniqueSongs, song)
		}
	}

	return uniqueSongs[:10], nil
}
