package main

import (
	"context"

	"github.com/google/go-github/github"
)

var publicReposFilter = github.RepositoryListByOrgOptions{Type: "public"}
var userPublicReposFilter = github.RepositoryListOptions{Type: "public"}

var openIssuesFilter = github.IssueListByRepoOptions{State: "open"}

func (syncer *Syncer) reposToSync() ([]*github.Repository, error) {
	options := publicReposFilter
	personal_options := userPublicReposFilter

	var repos []*github.Repository

	for {
		// if PersonalRepo flag is set to 'Y' then
		if syncer.PersonalRepo == "Y" {
			resources, resp, err := syncer.GithubClient.Repositories.List(
				context.TODO(),
				syncer.OrganizationName,
				&personal_options,
			)
			if err != nil {
				return nil, err
			}

			if len(resources) == 0 {
				break
			}

			for _, repo := range resources {
				if syncer.shouldSync(repo) {
					repos = append(repos, repo)
				}
			}

			if resp.NextPage == 0 {
				break
			}

			personal_options.ListOptions.Page = resp.NextPage

		} else {
			resources, resp, err := syncer.GithubClient.Repositories.ListByOrg(
				context.TODO(),
				syncer.OrganizationName,
				&options,
			)
			if err != nil {
				return nil, err
			}

			if len(resources) == 0 {
				break
			}

			for _, repo := range resources {
				if syncer.shouldSync(repo) {
					repos = append(repos, repo)
				}
			}

			if resp.NextPage == 0 {
				break
			}

			options.ListOptions.Page = resp.NextPage
		}
	}

	return repos, nil
}

func (syncer *Syncer) shouldSync(repository *github.Repository) bool {
	if len(syncer.Repositories) == 0 {
		return true
	}

	for _, name := range syncer.Repositories {
		if name == *repository.Name {
			return true
		}
	}

	return false
}

func (syncer *Syncer) allIssues(repo *github.Repository) ([]*github.Issue, error) {
	options := openIssuesFilter

	var all []*github.Issue

	for {
		resources, resp, err := syncer.GithubClient.Issues.ListByRepo(
			context.TODO(),
			*repo.Owner.Login,
			*repo.Name,
			&options,
		)
		if err != nil {
			return nil, err
		}

		if len(resources) == 0 {
			break
		}

		all = append(all, resources...)

		if resp.NextPage == 0 {
			break
		}

		options.ListOptions.Page = resp.NextPage
	}

	return all, nil
}

func (syncer *Syncer) allCommentsForIssue(
	repo *github.Repository,
	issue *github.Issue,
) ([]*github.IssueComment, error) {
	options := &github.IssueListCommentsOptions{}

	var all []*github.IssueComment

	for {
		resources, resp, err := syncer.GithubClient.Issues.ListComments(
			context.TODO(),
			*repo.Owner.Login,
			*repo.Name,
			*issue.Number,
			options,
		)
		if err != nil {
			return nil, err
		}

		if len(resources) == 0 {
			break
		}

		all = append(all, resources...)

		if resp.NextPage == 0 {
			break
		}

		options.ListOptions.Page = resp.NextPage
	}

	return all, nil
}
