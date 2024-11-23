package applemusic

import (
	"strconv"
	"strings"
)

type song struct {
	Track            string   `json:"track"`
	Artist           string   `json:"artist"`
	Album            string   `json:"album"`
	Genres           []string `json:"genres"`
	ReleaseDate      string   `json:"release_date"`
	DurationInMillis int      `json:"duration_in_millis"`
	AlbumArtURL      string   `json:"album_art_url"`
	URL              string   `json:"url"`
}

type songResponse struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Href       string `json:"href"`
	Attributes struct {
		AlbumName        string   `json:"albumName"`
		GenreNames       []string `json:"genreNames"`
		TrackNumber      int      `json:"trackNumber"`
		ReleaseDate      string   `json:"releaseDate"`
		DurationInMillis int      `json:"durationInMillis"`
		Artwork          struct {
			Width  int    `json:"width"`
			Height int    `json:"height"`
			URL    string `json:"url"`
		} `json:"artwork"`
		URL        string `json:"url"`
		Name       string `json:"name"`
		ArtistName string `json:"artistName"`
	} `json:"attributes"`
}

func songFromSongResponse(s songResponse) song {
	return song{
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
	}
}
