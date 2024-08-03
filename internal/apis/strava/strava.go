package strava

import (
	"github.com/gleich/lcp-v2/internal/cache"
	"github.com/gleich/lcp-v2/internal/secrets"
	"github.com/gleich/lumber/v2"
	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func Setup(router *chi.Mux) {
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
	stravaActivities := fetchActivities(*minioClient, stravaTokens)
	stravaCache := cache.NewCache("strava", stravaActivities)
	router.Get("/strava/cache", stravaCache.ServeHTTP())
	router.Post("/strava/event", eventRoute(stravaCache, *minioClient, stravaTokens))
	router.Get("/strava/event", challengeRoute)

	lumber.Success("setup strava cache")
}
