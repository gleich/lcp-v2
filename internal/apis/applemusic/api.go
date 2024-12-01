package applemusic

import (
	"net/http"
	"net/url"

	"github.com/gleich/lcp-v2/internal/apis"
	"github.com/gleich/lcp-v2/internal/secrets"
	"github.com/gleich/lumber/v3"
)

func sendAppleMusicAPIRequest[T any](path string) (T, error) {
	var zeroValue T
	u, err := url.JoinPath("https://api.music.apple.com/", path)
	if err != nil {
		lumber.Error(err, "failed to join URL")
		return zeroValue, err
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		lumber.Error(err, "failed to create request")
		return zeroValue, err
	}
	req.Header.Set("Authorization", "Bearer "+secrets.SECRETS.AppleMusicAppToken)
	req.Header.Set("Music-User-Token", secrets.SECRETS.AppleMusicUserToken)

	resp, err := apis.SendRequest[T](req)
	if err != nil {
		lumber.Error(err, "failed to make apple music API request")
		return zeroValue, err
	}
	return resp, nil
}
