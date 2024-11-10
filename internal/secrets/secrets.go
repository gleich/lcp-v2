package secrets

import (
	"github.com/caarlos0/env/v11"
	"github.com/gleich/lumber/v3"
	"github.com/joho/godotenv"
)

var SECRETS Secrets

type Secrets struct {
	CacheFolder string `env:"CACHE_FOLDER"`
	ValidToken  string `env:"VALID_TOKEN"`

	StravaClientID       string `env:"STRAVA_CLIENT_ID"`
	StravaClientSecret   string `env:"STRAVA_CLIENT_SECRET"`
	StravaOAuthCode      string `env:"STRAVA_OAUTH_CODE"`
	StravaAccessToken    string `env:"STRAVA_ACCESS_TOKEN"`
	StravaRefreshToken   string `env:"STRAVA_REFRESH_TOKEN"`
	StravaSubscriptionID int64  `env:"STRAVA_SUBSCRIPTION_ID"`
	StravaVerifyToken    string `env:"STRAVA_VERIFY_TOKEN"`
	MapboxAccessToken    string `env:"MAPBOX_ACCESS_TOKEN"`
	MinioEndpoint        string `env:"MINIO_ENDPOINT"`
	MinioAccessKeyID     string `env:"MINIO_ACCESS_KEY_ID"`
	MinioSecretKey       string `env:"MINIO_SECRET_KEY"`

	SteamKey string `env:"STEAM_KEY"`
	SteamID  string `env:"STEAM_ID"`

	GitHubAccessToken string `env:"GITHUB_ACCESS_TOKEN"`
}

func Load() {
	err := godotenv.Load()
	if err != nil {
		lumber.Fatal(err, "loading .env file failed")
	}
	loadedSecrets, err := env.ParseAs[Secrets]()
	if err != nil {
		lumber.Fatal(err, "parsing required env vars failed")
	}
	SECRETS = loadedSecrets
	lumber.Done("loaded secrets")
}
