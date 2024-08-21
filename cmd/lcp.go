package main

import (
	"net/http"
	"time"

	"github.com/gleich/lcp-v2/internal/apis/github"
	"github.com/gleich/lcp-v2/internal/apis/steam"
	"github.com/gleich/lcp-v2/internal/apis/strava"
	"github.com/gleich/lcp-v2/internal/secrets"
	"github.com/gleich/lumber/v3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	setupLogger()
	lumber.Info("booted")

	secrets.Load()

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RedirectSlashes)
	r.HandleFunc("/", rootRedirect)
	r.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	github.Setup(r)
	strava.Setup(r)
	steam.Setup(r)

	lumber.Info("starting server")
	err := http.ListenAndServe(":8000", r)
	if err != nil {
		lumber.Fatal(err, "failed to start router")
	}
}

func setupLogger() {
	nytime, err := time.LoadLocation("America/New_York")
	if err != nil {
		lumber.Fatal(err, "failed to load new york timezone")
	}
	lumber.SetTimezone(nytime)
	lumber.SetTimeFormat("01/02/2006 03:04:05 PM MST")
}

func rootRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://mattglei.ch/lcp", http.StatusTemporaryRedirect)
}
