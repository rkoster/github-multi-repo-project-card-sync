package github

import (
	"context"
	"strings"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Client struct {
	client *githubv4.Client
}

func NewClient(token string, ctx context.Context) *Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)

	return &Client{githubv4.NewClient(httpClient)}
}

func (c *Client) GetOrganizationProject(org string, projectNumber int, ctx context.Context) (Project, error) {
	var q struct {
		Organization struct {
			ProjectNext struct {
				ID     string
				Fields struct {
					Nodes    []ProjectField
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"fields(first: 100, after: $fieldsCursor)"`
			} `graphql:"projectNext(number: $projectNumber)"`
		} `graphql:"organization(login: $organization)"`
	}

	variables := map[string]interface{}{
		"organization":  githubv4.String(org),
		"projectNumber": githubv4.Int(projectNumber),
		"fieldsCursor":  (*githubv4.String)(nil), // Null after argument to get first page.
	}

	var project Project
	project.Fields = make([]ProjectField, 0)

	for {
		err := c.client.Query(ctx, &q, variables)
		if err != nil {
			return project, err
		}
		project.ID = q.Organization.ProjectNext.ID
		project.Fields = append(project.Fields, q.Organization.ProjectNext.Fields.Nodes...)
		if !q.Organization.ProjectNext.Fields.PageInfo.HasNextPage {
			break
		}
		variables["fieldsCursor"] = githubv4.NewString(q.Organization.ProjectNext.Fields.PageInfo.EndCursor)
	}

	return project, nil
}

func (c *Client) ListOpenPullRequests(repo string, ctx context.Context) ([]PullRequest, error) {
	owner := strings.Split(repo, "/")[0]
	name := strings.Split(repo, "/")[1]

	var q struct {
		Repository struct {
			PullRequests struct {
				Nodes    []PullRequest
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"pullRequests(first: 100, after: $pullRequestsCursor, states: [OPEN])"` // 100 per page.
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}

	variables := map[string]interface{}{
		"repositoryOwner":    githubv4.String(owner),
		"repositoryName":     githubv4.String(name),
		"pullRequestsCursor": (*githubv4.String)(nil), // Null after argument to get first page.
	}

	// Get pullRequests from all pages.
	var allPullRequests []PullRequest
	for {
		err := c.client.Query(ctx, &q, variables)
		if err != nil {
			return []PullRequest{}, err
		}
		allPullRequests = append(allPullRequests, q.Repository.PullRequests.Nodes...)
		if !q.Repository.PullRequests.PageInfo.HasNextPage {
			break
		}
		variables["pullRequestsCursor"] = githubv4.NewString(q.Repository.PullRequests.PageInfo.EndCursor)
	}

	return allPullRequests, nil
}
