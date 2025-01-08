package strava

import (
	"net/http"

	"github.com/gleich/lumber/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"pkg.mattglei.ch/lcp-v2/internal/cache"
	"pkg.mattglei.ch/lcp-v2/internal/secrets"
)

func Setup(mux *http.ServeMux) {
	stravaTokens := loadTokens()
	stravaTokens.refreshIfNeeded()
	minioClient, err := minio.New(secrets.SECRETS.MinioEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			secrets.SECRETS.MinioAccessKeyID,
			secrets.SECRETS.MinioSecretKey,
			"",
		),
		Secure: true,
	})
	if err != nil {
		lumber.Fatal(err, "failed to create minio client")
	}
	stravaActivities, err := fetchActivities(*minioClient, stravaTokens)
	if err != nil {
		lumber.ErrorMsg("failed to load initial data for strava cache; not updating")
	}
	stravaCache := cache.New("strava", stravaActivities)
	mux.HandleFunc("GET /strava", stravaCache.ServeHTTP)
	mux.HandleFunc("POST /strava/event", eventRoute(stravaCache, *minioClient, stravaTokens))
	mux.HandleFunc("GET /strava/event", challengeRoute)

	lumber.Done("setup strava cache")
}
