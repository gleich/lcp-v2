package github

import (
	"context"
	"time"

	"github.com/gleich/lcp-v2/pkg/cache"
	"github.com/gleich/lcp-v2/pkg/secrets"
	"github.com/gleich/lumber/v2"
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
	githubCache := cache.NewCache("github", FetchPinnedRepos(githubClient))
	router.Get("/github/cache", githubCache.ServeHTTP())
	go githubCache.StartPeriodicUpdate(func() []repository { return FetchPinnedRepos(githubClient) }, 2*time.Minute)
	lumber.Success("setup github cache")
}
