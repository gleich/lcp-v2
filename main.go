package main

import (
	"net/http"

	"github.com/caarlos0/env/v11"
	"github.com/gleich/lcp-v2/pkg/apis/steam"
	"github.com/gleich/lcp-v2/pkg/cache"
	"github.com/gleich/lcp-v2/pkg/secrets"
	"github.com/gleich/lumber/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	lumber.Info("booted")

	err := godotenv.Load()
	if err != nil {
		lumber.Fatal(err, "Error loading .env file")
	}
	loadedSecrets, err := env.ParseAs[secrets.SecretsData]()
	if err != nil {
		lumber.Fatal(err, "parsing required env vars failed")
	}
	secrets.SECRETS = loadedSecrets
	lumber.Success("loaded secrets")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.HandleFunc("/", rootRedirect)

	games := steam.FetchRecentlyPlayedGames()
	steamCache := cache.New("steam", games)
	r.Get("/steam/cache", steamCache.Route())
	lumber.Success("init steam cache")

	// stravaTokens := strava.LoadTokens()
	// stravaTokens.RefreshIfNeeded()
	// stravaActivities := strava.FetchActivities(stravaTokens)
	// stravaCache := cache.New("strava", stravaActivities)
	// r.Get("/strava/cache", stravaCache.Route())
	// r.Post("/strava/event", strava.EventRoute(&stravaCache, stravaTokens))
	// r.Get("/strava/event", strava.ChallengeRoute)
	// lumber.Success("init strava cache")

	err = http.ListenAndServe(":8000", r)
	if err != nil {
		lumber.Fatal(err, "failed to start router")
	}
}

func rootRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://mattglei.ch", http.StatusTemporaryRedirect)
}
