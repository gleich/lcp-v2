package main

import (
	"context"
	"net/http"

	"github.com/caarlos0/env/v11"
	"github.com/gleich/lcp-v2/pkg/apis/github"
	"github.com/gleich/lcp-v2/pkg/apis/steam"
	"github.com/gleich/lcp-v2/pkg/apis/strava"
	"github.com/gleich/lcp-v2/pkg/cache"
	"github.com/gleich/lcp-v2/pkg/secrets"
	"github.com/gleich/lumber/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
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

	githubTokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: secrets.SECRETS.GitHubAccessToken},
	)
	githubHttpClient := oauth2.NewClient(context.Background(), githubTokenSource)
	githubClient := githubv4.NewClient(githubHttpClient)
	githubCache := cache.New("github", github.FetchPinnedRepos(githubClient))
	r.Get("/github/cache", githubCache.Route())
	lumber.Success("init github cache")

	stravaTokens := strava.LoadTokens()
	stravaTokens.RefreshIfNeeded()
	stravaActivities := strava.FetchActivities(stravaTokens)
	stravaCache := cache.New("strava", stravaActivities)
	r.Get("/strava/cache", stravaCache.Route())
	r.Post("/strava/event", strava.EventRoute(&stravaCache, stravaTokens))
	r.Get("/strava/event", strava.ChallengeRoute)
	lumber.Success("init strava cache")

	games := steam.FetchRecentlyPlayedGames()
	steamCache := cache.New("steam", games)
	r.Get("/steam/cache", steamCache.Route())
	lumber.Success("init steam cache")

	err = http.ListenAndServe(":8000", r)
	if err != nil {
		lumber.Fatal(err, "failed to start router")
	}
}

func rootRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://mattglei.ch", http.StatusTemporaryRedirect)
}
