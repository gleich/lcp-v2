package strava

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gleich/lcp-v2/internal/cache"
	"github.com/gleich/lcp-v2/internal/secrets"
	"github.com/gleich/lumber/v2"
	"github.com/minio/minio-go/v7"
)

type event struct {
	AspectType     string            `json:"aspect_type"`
	EventTime      int64             `json:"event_time"`
	ObjectID       int64             `json:"object_id"`
	ObjectType     string            `json:"object_type"`
	OwnerID        int64             `json:"owner_id"`
	SubscriptionID int64             `json:"subscription_id"`
	Updates        map[string]string `json:"updates"`
}

func eventRoute(
	stravaCache *cache.Cache[[]activity],
	minioClient minio.Client,
	tokens tokens,
) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			lumber.Error(err, "reading response body failed")
			return
		}

		var eventData event
		err = json.Unmarshal(body, &eventData)
		if err != nil {
			lumber.Error(err, "failed to parse json")
			lumber.Debug(string(body))
			return
		}

		if eventData.SubscriptionID != secrets.SECRETS.StravaSubscriptionID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokens.refreshIfNeeded()
		stravaCache.Update(fetchActivities(minioClient, tokens))
	})
}

func challengeRoute(w http.ResponseWriter, r *http.Request) {
	verifyToken := r.URL.Query().Get("hub.verify_token")
	if verifyToken != secrets.SECRETS.StravaVerifyToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	challenge := r.URL.Query().Get("hub.challenge")
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(struct {
		Challenge string `json:"hub.challenge"`
	}{Challenge: challenge})
	if err != nil {
		lumber.Error(err, "failed to write json")
	}
}
