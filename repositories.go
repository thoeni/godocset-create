package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/v24/github"
)

type OrganizationRepository struct {
	name    string
	page    int
	perPage int
}

func (o OrganizationRepository) Total(ctx context.Context, client *github.Client) int {
	org, _, err := client.Organizations.Get(ctx, o.name)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	return org.GetPublicRepos() + org.GetTotalPrivateRepos()
}

func (o OrganizationRepository) Next() Lister {
	res := o
	res.page++
	return res
}

func (o OrganizationRepository) List(ctx context.Context, client *github.Client) ([]*github.Repository, *github.Response, error) {
	return client.Repositories.ListByOrg(ctx, o.name, &github.RepositoryListByOrgOptions{
		Type: "all",
		ListOptions: github.ListOptions{
			Page:    o.page,
			PerPage: o.perPage,
		},
	})
}

type UserRepository struct {
	name    string
	page    int
	perPage int
}

func (u UserRepository) Total(ctx context.Context, client *github.Client) int {
	user, _, err := client.Users.Get(ctx, u.name)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	return user.GetPublicRepos() + user.GetTotalPrivateRepos()
}

func (u UserRepository) Next() Lister {
	res := u
	res.page++
	return res
}

func (u UserRepository) List(ctx context.Context, client *github.Client) ([]*github.Repository, *github.Response, error) {
	return client.Repositories.List(ctx, u.name, &github.RepositoryListOptions{
		Type: "all",
		ListOptions: github.ListOptions{
			Page:    u.page,
			PerPage: u.perPage,
		},
	})
}
