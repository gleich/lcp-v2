package github

import (
	"context"
	"net/http"
	"time"

	"github.com/gleich/lumber/v3"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"pkg.mattglei.ch/lcp-v2/internal/cache"
	"pkg.mattglei.ch/lcp-v2/internal/secrets"
)

func Setup(mux *http.ServeMux) {
	githubTokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: secrets.SECRETS.GitHubAccessToken},
	)
	githubHttpClient := oauth2.NewClient(context.Background(), githubTokenSource)
	githubClient := githubv4.NewClient(githubHttpClient)

	pinnedRepos, err := fetchPinnedRepos(githubClient)
	if err != nil {
		lumber.Error(err, "fetching initial pinned repos failed")
	}

	githubCache := cache.New("github", pinnedRepos, err == nil)
	mux.HandleFunc("GET /github", githubCache.ServeHTTP)
	go githubCache.UpdatePeriodically(
		func() ([]repository, error) { return fetchPinnedRepos(githubClient) },
		1*time.Minute,
	)
	lumber.Done("setup github cache")
}
