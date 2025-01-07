package apis

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gleich/lumber/v3"
)

var WarningError = errors.New("Warning error when trying to make request. Ignore error.")

// sends a given http.Request and will unmarshal the JSON from the response body and return that as the given type.
func SendRequest[T any](req *http.Request) (T, error) {
	var zeroValue T // to be used as "nil" when returning errors
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
		lumber.Warning(
			"status code of",
			resp.StatusCode,
			"returned from API. Code of 200 expected from",
			req.URL.String(),
		)
		return zeroValue, WarningError
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
