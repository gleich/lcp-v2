package strava

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gleich/lcp-v2/internal/apis"
	"github.com/gleich/lumber/v3"
)

func sendStravaAPIRequest[T any](path string, tokens tokens) (T, error) {
	var zeroValue T

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://www.strava.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		lumber.Error(err, "failed to create request")
		return zeroValue, err
	}
	req.Header.Set("Authorization", "Bearer "+tokens.Access)

	resp, err := apis.SendRequest[T](req)
	if err != nil {
		if !errors.Is(err, apis.WarningError) {
			lumber.Error(err, "failed to make strava API request")
		}
		return zeroValue, err
	}
	return resp, nil
}
