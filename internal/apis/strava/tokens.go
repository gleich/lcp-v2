package strava

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gleich/lcp-v2/internal/apis"
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
	// subtract 60 to ensure that token doesn't expire in the next 60 seconds
	if t.ExpiresAt-60 >= time.Now().Unix() {
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

	tokens, err := apis.SendRequest[tokens](req)
	if err != nil {
		lumber.Error(err, "failed to refresh tokens")
		return
	}

	*t = tokens
	lumber.Done("loaded new strava access token:", t.Access)
}
