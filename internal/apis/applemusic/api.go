package applemusic

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gleich/lumber/v3"
	"pkg.mattglei.ch/lcp-2/internal/apis"
	"pkg.mattglei.ch/lcp-2/internal/secrets"
)

func sendAppleMusicAPIRequest[T any](path string) (T, error) {
	var zeroValue T
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://api.music.apple.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		lumber.Error(err, "failed to create request")
		return zeroValue, err
	}
	req.Header.Set("Authorization", "Bearer "+secrets.SECRETS.AppleMusicAppToken)
	req.Header.Set("Music-User-Token", secrets.SECRETS.AppleMusicUserToken)

	resp, err := apis.SendRequest[T](req)
	if err != nil {
		if !errors.Is(err, apis.WarningError) {
			lumber.Error(err, "failed to make apple music API request")
		}
		return zeroValue, err
	}
	return resp, nil
}
