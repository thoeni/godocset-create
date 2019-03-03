package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"sync"

	"github.com/google/go-github/v24/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

func main() {

	const perPage = 100
	const maxConcurrentUpdates = 10

	githubToken := flag.String("githubToken", "", "personal access token allowed to pull repos")
	organization := flag.String("organization", "deliveroo", "organization to pull docs form")
	filter := flag.String("filter", "", "filter to specify a subset of the org repos to pull down (works in a 'contains' fashion)")

	flag.Parse()
	fmt.Println("Using token=", *githubToken, " and org=", *organization)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	org, _, err := client.Organizations.Get(ctx, *organization)
	if err != nil {
		fmt.Println(err)
		return
	}

	totalRepos := org.GetPublicRepos() + org.GetTotalPrivateRepos()
	fmt.Println("Owned repos", totalRepos)
	concurrency := totalRepos / perPage
	if totalRepos-concurrency*perPage > 0 {
		concurrency++
	}

	var wg sync.WaitGroup
	goRepoMetadataCh := make(chan *github.Repository, totalRepos)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go fetchGithubRepoMetadata(ctx, &wg, client, *organization, *filter, i+1, perPage, goRepoMetadataCh)
	}

	go func() {
		wg.Wait()
		close(goRepoMetadataCh)
	}()

	concurrentUpdateLimiter := make(chan struct{}, maxConcurrentUpdates)
	var cloneWg sync.WaitGroup
	var count int
	for r := range goRepoMetadataCh {
		fmt.Printf("Repo: %s\n", r.GetName())
		fmt.Printf("CloneURL: %s\n", r.GetCloneURL())
		fmt.Printf("Ready to clone...\n\n")
		concurrentUpdateLimiter <- struct{}{}
		cloneWg.Add(1)
		go updateRepo(ctx, &cloneWg, concurrentUpdateLimiter, r.GetCloneURL(), r.GetName(), *githubToken, *organization)
		count++
	}

	cloneWg.Wait()
	fmt.Printf("Fetched %d Go repositories.\n", count)
}

func updateRepo(ctx context.Context, wg *sync.WaitGroup, w <-chan struct{}, url, repoName, githubToken, organization string) {
	defer wg.Done()
	defer func(){
		<-w
	}()
	repoPath := fmt.Sprintf("/go/src/github.com/%s/%s", organization, repoName)
	var r *git.Repository
	// Attempt opening the repo
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		switch err {
		case git.ErrRepositoryNotExists:
			// Clone the repo if doesn't exist
			if _, err := git.PlainCloneContext(ctx, repoPath, false, &git.CloneOptions{
				URL: url,
				Auth: &http.BasicAuth{
					Username: organization,
					Password: githubToken,
				},
			}); err != nil {
				fmt.Println(errors.Wrap(err, "cannot clone repository"))
			}
			return
		default:
			fmt.Println(errors.Wrap(err, "cannot open repository"))
			return
		}
	}

	if err := r.FetchContext(ctx, &git.FetchOptions{
		RemoteName: "origin",
		Auth: &http.BasicAuth{
			Username: organization,
			Password: githubToken,
		},
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		fmt.Println(errors.Wrap(err, "cannot fetch repository"))
	}
}

func fetchGithubRepoMetadata(ctx context.Context, wg *sync.WaitGroup, client *github.Client, organization, filter string, page, perPage int, r chan<- *github.Repository) {
	defer wg.Done()
	opt := &github.RepositoryListByOrgOptions{
		Type: "all",
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}

	repos, _, err := client.Repositories.ListByOrg(ctx, organization, opt)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := range repos {
		if repos[i].GetLanguage() == "Go" && strings.Contains(repos[i].GetCloneURL(), filter) {
			r <- repos[i]
		}
	}
}
