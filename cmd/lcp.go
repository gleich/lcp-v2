package main

import (
	"net/http"
	"time"

	"github.com/gleich/lumber/v3"
	"pkg.mattglei.ch/lcp-2/internal/apis/applemusic"
	"pkg.mattglei.ch/lcp-2/internal/apis/github"
	"pkg.mattglei.ch/lcp-2/internal/apis/steam"
	"pkg.mattglei.ch/lcp-2/internal/apis/strava"
	"pkg.mattglei.ch/lcp-2/internal/secrets"
)

func main() {
	setupLogger()
	lumber.Info("booted")

	secrets.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootRedirect)

	github.Setup(mux)
	strava.Setup(mux)
	steam.Setup(mux)
	applemusic.Setup(mux)

	lumber.Info("starting server")
	err := http.ListenAndServe(":8000", mux)
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
	lumber.SetTimeFormat("01/02 03:04:05 PM MST")
}

func rootRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://mattglei.ch/lcp", http.StatusTemporaryRedirect)
}
