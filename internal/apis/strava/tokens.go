package strava

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gleich/lcp-v2/internal/secrets"
	"github.com/gleich/lumber/v3"
)

type tokens struct {
	Access    string `json:"access_token"`
	Refresh   string `json:"refresh_token"`
	ExpiresAt int64  `json:"expires_at"`
}

func loadTokens() tokens {
	return tokens{
		Access:    secrets.SECRETS.StravaAccessToken,
		Refresh:   secrets.SECRETS.StravaRefreshToken,
		ExpiresAt: 0, // starts at zero to force a refresh on boot
	}
}

func (t *tokens) refreshIfNeeded() {
	if t.ExpiresAt >= time.Now().Unix() {
		return
	}

	params := url.Values{
		"client_id":     {secrets.SECRETS.StravaClientID},
		"client_secret": {secrets.SECRETS.StravaClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {t.Refresh},
		"code":          {secrets.SECRETS.StravaOAuthCode},
	}
	req, err := http.NewRequest("POST", "https://www.strava.com/oauth/token?"+params.Encode(), nil)
	if err != nil {
		lumber.Error(err, "creating request for new token failed")
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		lumber.Error(err, "sending request for new data token failed")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		lumber.Error(err, "reading response body failed")
		return
	}
	if resp.StatusCode != http.StatusOK {
		lumber.ErrorMsg(resp.StatusCode, "when trying to get new token data:", string(body))
		return
	}

	var tokens tokens
	err = json.Unmarshal(body, &tokens)
	if err != nil {
		lumber.Error(err, "failed to parse json")
		lumber.Debug("body:", string(body))
		return
	}

	os.Setenv("STRAVA_ACCESS_TOKEN", tokens.Access)
	os.Setenv("STRAVA_REFRESH_TOKEN", tokens.Refresh)
	os.Setenv("STRAVA_REFRESH_TOKEN_EXPIRATION", strconv.FormatInt(tokens.ExpiresAt, 10))
	*t = tokens

	lumber.Done("loaded new strava access token:", t.Access)
}
