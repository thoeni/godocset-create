package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/google/go-github/v24/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

func main() {

	const perPage = 100
	const maxConcurrentClone = 10

	var githubToken = os.Getenv("GITHUB_TOKEN")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	org, _, err := client.Organizations.Get(ctx, "deliveroo")
	if err != nil {
		fmt.Println(err)
		return
	}

	totalRepos := org.GetOwnedPrivateRepos()
	fmt.Println("Owned private repos", totalRepos)
	concurrency := totalRepos / perPage
	if totalRepos-concurrency*perPage > 0 {
		concurrency++
	}

	var wg sync.WaitGroup
	goRepoCh := make(chan *github.Repository, totalRepos)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go fetchRepos(ctx, &wg, client, i+1, perPage, goRepoCh)
	}

	go func() {
		wg.Wait()
		close(goRepoCh)
	}()

	concurrentCloneLimiter := make(chan struct{}, maxConcurrentClone)
	for r := range goRepoCh {
		fmt.Printf("Repo: %s\n", r.GetName())
		fmt.Printf("CloneURL: %s\n", r.GetCloneURL())
		fmt.Printf("Ready to clone...\n\n")
		concurrentCloneLimiter <- struct{}{}
		go cloneRepo(ctx, concurrentCloneLimiter, r.GetCloneURL(), r.GetName(), githubToken)
	}

	fmt.Println("Done.")
}

func cloneRepo(ctx context.Context, w <-chan struct{}, url, repoName, githubToken string) {
	if _, err := git.PlainCloneContext(ctx, fmt.Sprintf("/tmp/deliveroo/%s", repoName), false, &git.CloneOptions{
		URL: url,
		Auth: &http.BasicAuth{
			Username: "deliveroo",
			Password: githubToken,
		},
	}); err != nil {
		fmt.Println("Something went wrong:", err)
	}
	<-w
}

func fetchRepos(ctx context.Context, wg *sync.WaitGroup, client *github.Client, page, perPage int, r chan<- *github.Repository) {
	defer wg.Done()
	opt := &github.RepositoryListByOrgOptions{
		Type: "private",
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}

	repos, _, err := client.Repositories.ListByOrg(ctx, "deliveroo", opt)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := range repos {
		if repos[i].GetLanguage() == "Go" {
			r <- repos[i]
		}
	}
}
