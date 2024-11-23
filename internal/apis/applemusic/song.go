package applemusic

type song struct {
	Track            string   `json:"track"`
	Artist           string   `json:"artist"`
	Album            string   `json:"album"`
	Genres           []string `json:"genres"`
	ReleaseDate      string   `json:"releaseDate"`
	DurationInMillis int      `json:"durationInMillis"`
	AlbumArtURL      string   `json:"albumArtURL"`
	URL              string   `json:"url"`
}

type songListData []struct {
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
