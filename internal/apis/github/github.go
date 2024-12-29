package github

import (
	"context"
	"time"

	"github.com/gleich/lcp-v2/internal/cache"
	"github.com/gleich/lcp-v2/internal/secrets"
	"github.com/gleich/lumber/v3"
	"github.com/go-chi/chi/v5"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
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
	router.Get("/github/cache", githubCache.ServeHTTP())
	router.Handle("/github/cache/ws", githubCache.ServeWS())
	go githubCache.UpdatePeriodically(
		func() ([]repository, error) { return fetchPinnedRepos(githubClient) },
		2*time.Minute,
	)
	lumber.Done("setup github cache")
}
