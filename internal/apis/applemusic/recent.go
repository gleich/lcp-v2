package applemusic

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gleich/lumber/v3"
)

type recentlyPlayedResponse struct {
	Next string       `json:"next"`
	Data songListData `json:"data"`
}

func FetchRecentlyPlayed() {
	response, err := sendAPIRequest[recentlyPlayedResponse](
		"https://api.music.apple.com/v1/me/recent/played/tracks",
	)
	if err != nil {
		lumber.Fatal(err, "failed to send request for recently played songs")
	}

	var songs []song
	for _, s := range response.Data {
		songs = append(songs, song{
			Track:            s.Attributes.Name,
			Artist:           s.Attributes.ArtistName,
			Album:            s.Attributes.AlbumName,
			Genres:           s.Attributes.GenreNames,
			ReleaseDate:      s.Attributes.ReleaseDate,
			DurationInMillis: s.Attributes.DurationInMillis,
			AlbumArtURL: strings.ReplaceAll(strings.ReplaceAll(
				s.Attributes.Artwork.URL,
				"{w}",
				strconv.Itoa(s.Attributes.Artwork.Width),
			), "{h}", strconv.Itoa(s.Attributes.Artwork.Height)),
			URL: s.Attributes.URL,
		})
	}

	encodedData, _ := json.Marshal(songs)
	lumber.Debug(string(encodedData))
}
