package github

import (
	"context"
	"time"

	"github.com/gleich/lumber/v3"
	"github.com/go-chi/chi/v5"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"pkg.mattglei.ch/lcp-v2/internal/cache"
	"pkg.mattglei.ch/lcp-v2/internal/secrets"
)

func Setup(router *chi.Mux) {
	githubTokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: secrets.SECRETS.GitHubAccessToken},
	)
	githubHttpClient := oauth2.NewClient(context.Background(), githubTokenSource)
	githubClient := githubv4.NewClient(githubHttpClient)

	pinnedRepos, err := fetchPinnedRepos(githubClient)
	if err != nil {
		lumber.Fatal(err, "fetching initial pinned repos failed")
	}

	githubCache := cache.New("github", pinnedRepos)
	router.Get("/github", githubCache.ServeHTTP())
	go githubCache.UpdatePeriodically(
		func() ([]repository, error) { return fetchPinnedRepos(githubClient) },
		1*time.Minute,
	)
	lumber.Done("setup github cache")
}
