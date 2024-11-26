package applemusic

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gleich/lcp-v2/internal/secrets"
	"github.com/gleich/lumber/v3"
)

func sendAPIRequest[T any](endpoint string) (T, error) {
	var zeroValue T // to be used as "nil" when returning errors
	req, err := http.NewRequest("GET", "https://api.music.apple.com/"+endpoint, nil)
	if err != nil {
		lumber.Error(err, "creating request failed")
		return zeroValue, err
	}
	req.Header.Set("Authorization", "Bearer "+secrets.SECRETS.AppleMusicAppToken)
	req.Header.Set("Music-User-Token", secrets.SECRETS.AppleMusicUserToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		lumber.Error(err, "sending request failed")
		return zeroValue, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		lumber.Error(err, "reading response body failed")
		return zeroValue, err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf(
			"status code of %d returned from apple music API. Code of 200 expected",
			resp.StatusCode,
		)
		if resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusGatewayTimeout ||
			resp.StatusCode == http.StatusInternalServerError {
			lumber.Warning(err)
		} else {
			lumber.Error(err)
		}
		return zeroValue, err
	}

	var data T
	err = json.Unmarshal(body, &data)
	if err != nil {
		lumber.Error(err, "failed to parse json")
		lumber.Debug(string(body))
		return zeroValue, err
	}

	return data, nil
}
