package secrets

var SECRETS SecretsData

type SecretsData struct {
	ValidToken string `env:"VALID_TOKEN"`

	StravaClientID               string `env:"STRAVA_CLIENT_ID"`
	StravaClientSecret           string `env:"STRAVA_CLIENT_SECRET"`
	StravaOAuthCode              string `env:"STRAVA_OAUTH_CODE"`
	StravaAccessToken            string `env:"STRAVA_ACCESS_TOKEN"`
	StravaRefreshToken           string `env:"STRAVA_REFRESH_TOKEN"`
	StravaRefreshTokenExpiration int64  `env:"STRAVA_REFRESH_TOKEN_EXPIRATION"`
	StravaSubscriptionID         int64  `env:"STRAVA_SUBSCRIPTION_ID"`
	StravaVerifyToken            string `env:"STRAVA_VERIFY_TOKEN"`

	SteamKey string `env:"STEAM_KEY"`
	SteamID  string `env:"STEAM_ID"`

	GitHubAccessToken string `env:"GITHUB_ACCESS_TOKEN"`
}
