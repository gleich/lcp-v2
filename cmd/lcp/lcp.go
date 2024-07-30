package main

import (
	"net/http"

	"github.com/gleich/lcp-v2/internal/apis/github"
	"github.com/gleich/lcp-v2/internal/apis/steam"
	"github.com/gleich/lcp-v2/internal/apis/strava"
	"github.com/gleich/lcp-v2/internal/secrets"
	"github.com/gleich/lumber/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
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

func rootRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://mattglei.ch", http.StatusTemporaryRedirect)
}
