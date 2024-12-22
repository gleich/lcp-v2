package strava

import (
	"github.com/gleich/lcp-v2/internal/cache"
	"github.com/gleich/lcp-v2/internal/secrets"
	"github.com/gleich/lumber/v3"
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
	stravaActivities, err := fetchActivities(*minioClient, stravaTokens)
	if err != nil {
		lumber.ErrorMsg("failed to load initial data for strava cache; not updating")
	}
	stravaCache := cache.NewCache("strava", stravaActivities)
	router.Get("/strava/cache", stravaCache.ServeHTTP())
	router.Handle("/strava/cache/ws", stravaCache.ServeWS())
	router.Post("/strava/event", eventRoute(stravaCache, *minioClient, stravaTokens))
	router.Get("/strava/event", challengeRoute)

	lumber.Done("setup strava cache")
}
