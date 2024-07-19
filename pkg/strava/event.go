package strava

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gleich/lcp-v2/pkg/cache"
	"github.com/gleich/lcp-v2/pkg/secrets"
	"github.com/gleich/lumber/v2"
)

type Event struct {
	AspectType     string            `json:"aspect_type"`
	EventTime      int64             `json:"event_time"`
	ObjectID       int64             `json:"object_id"`
	ObjectType     string            `json:"object_type"`
	OwnerID        int64             `json:"owner_id"`
	SubscriptionID int64             `json:"subscription_id"`
	Updates        map[string]string `json:"updates"`
}

func EventRoute(stravaCache *cache.Cache[[]Activity], tokens Tokens) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			lumber.Error(err, "reading response body failed")
			return
		}

		var event Event
		err = json.Unmarshal(body, &event)
		if err != nil {
			lumber.Error(err, "failed to parse json")
			lumber.Debug(string(body))
			return
		}

		if event.SubscriptionID != secrets.SECRETS.StravaSubscriptionID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokens.RefreshIfNeeded()
		stravaCache.Update(FetchActivities(tokens))
	})
}

func ChallengeRoute(w http.ResponseWriter, r *http.Request) {
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
