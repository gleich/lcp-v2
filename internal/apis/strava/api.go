package strava

import (
	"net/http"
	"net/url"

	"github.com/gleich/lcp-v2/internal/apis"
	"github.com/gleich/lumber/v3"
)

func sendStravaAPIRequest[T any](path string, tokens tokens) (T, error) {
	var zeroValue T
	u, err := url.JoinPath("https://www.strava.com/", path)
	if err != nil {
		lumber.Error(err, "failed to join URL")
		return zeroValue, err
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		lumber.Error(err, "failed to create request")
		return zeroValue, err
	}
	req.Header.Set("Authorization", "Bearer "+tokens.Access)

	resp, err := apis.SendRequest[T](req)
	if err != nil {
		lumber.Error(err, "failed to make strava API request")
		return zeroValue, err
	}
	return resp, nil
}
