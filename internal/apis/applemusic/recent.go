package applemusic

type recentlyPlayedResponse struct {
	Data []songResponse `json:"data"`
}

func fetchRecentlyPlayed() ([]song, error) {
	response, err := sendAPIRequest[recentlyPlayedResponse](
		"v1/me/recent/played/tracks",
	)
	if err != nil {
		return []song{}, err
	}

	var songs []song
	for _, s := range response.Data {
		songs = append(songs, songFromSongResponse(s))
	}
	return songs, nil
}
