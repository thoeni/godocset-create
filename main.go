package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v24/github"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"strings"
	"sync"
	"sync/atomic"
)

func main() {

	const perPage = 100
	const maxConcurrentUpdates = 10

	viper.SetConfigName("godocset-config")
	viper.AddConfigPath("/tmp")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	githubToken := viper.GetString("Github.token")
	githubUser := viper.GetString("Github.user_id")
	cloneTargetDir := viper.GetString("Github.clone_target_dir")
	organizations := viper.GetStringSlice("Docset.organizations")
	users := viper.GetStringSlice("Docset.users")
	filter := viper.GetStringSlice("Docset.filters")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	goRepoMetadataCh := make(chan *github.Repository, 2*maxConcurrentUpdates)
	var cloneWg sync.WaitGroup
	var count int32
	go func() {
		for r := range goRepoMetadataCh {
			cloneWg.Add(1)
			concurrentUpdateLimiter := make(chan struct{}, maxConcurrentUpdates)
			fmt.Printf("Repo: %s\n", r.GetName())
			fmt.Printf("CloneURL: %s\n", r.GetCloneURL())
			fmt.Printf("Ready to clone...\n\n")
			concurrentUpdateLimiter <- struct{}{}
			go updateRepo(ctx, &cloneWg, concurrentUpdateLimiter, r.GetCloneURL(), cloneTargetDir, r.GetFullName(), githubToken, githubUser)
			atomic.AddInt32(&count, 1)
		}
	}()

	var githubRepoPagesWaitGroup sync.WaitGroup

	repoListers := make([]Lister, 0, len(organizations)+len(users))
	for _, org := range organizations {
		repoListers = append(repoListers, OrganizationRepository{
			name:    org,
			perPage: perPage,
		})
	}
	for _, user := range users {
		repoListers = append(repoListers, UserRepository{
			name:    user,
			perPage: perPage,
		})
	}

	for _, repoLister := range repoListers {
		totalRepos := repoLister.Total(ctx, client)
		fmt.Println("Owned repos", totalRepos)
		concurrency := totalRepos / perPage
		if totalRepos-concurrency*perPage > 0 {
			concurrency++
		}

		for i := 0; i < concurrency; i++ {
			githubRepoPagesWaitGroup.Add(1)
			repoLister = repoLister.Next()
			go fetchGithubRepoMetadata(ctx, &githubRepoPagesWaitGroup, client, repoLister, filter, goRepoMetadataCh)
		}
	}

	githubRepoPagesWaitGroup.Wait()
	close(goRepoMetadataCh)
	cloneWg.Wait()
}

func updateRepo(ctx context.Context, wg *sync.WaitGroup, w <-chan struct{}, url, cloneTargetDir, repoName, githubToken, githubUser string) {
	defer func() {
		wg.Done()
		<-w
	}()
	repoPath := fmt.Sprintf("%s/src/github.com/%s", cloneTargetDir, repoName)
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
					Username: githubUser,
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
			Username: githubUser,
			Password: githubToken,
		},
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		fmt.Println(errors.Wrap(err, "cannot fetch repository"))
	}
}

type Lister interface {
	List(ctx context.Context, client *github.Client) ([]*github.Repository, *github.Response, error)
	Total(ctx context.Context, client *github.Client) int
	Next() Lister
}

func fetchGithubRepoMetadata(ctx context.Context, wg *sync.WaitGroup, client *github.Client, l Lister, filter []string, r chan<- *github.Repository) {
	defer wg.Done()
	repos, _, err := l.List(ctx, client)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := range repos {
		if repos[i].GetLanguage() == "Go" && matchFilter(repos[i].GetFullName(), filter) {
			r <- repos[i]
		}
	}
}

func matchFilter(keyword string, filter []string) bool {
	if len(filter) == 0 {
		return true
	}
	for _, f := range filter {
		if strings.Contains(f, keyword) {
			return true
		}
	}
	return false
}
