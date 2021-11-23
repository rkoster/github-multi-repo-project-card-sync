package github

import (
	"context"
	"fmt"
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
				ID     githubv4.ID
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

func (c *Client) AddProjectItem(projectID, itemID githubv4.ID, ctx context.Context) (ProjectItem, error) {
	var m struct {
		AddProjectNextItem struct {
			ProjectNextItem ProjectItem
		} `graphql:"addProjectNextItem(input: $input)"`
	}
	input := githubv4.AddProjectNextItemInput{
		ProjectID: projectID,
		ContentID: itemID,
	}

	err := c.client.Mutate(ctx, &m, input, nil)
	if err != nil {
		return ProjectItem{}, err
	}

	if m.AddProjectNextItem.ProjectNextItem.FieldValues.PageInfo.HasNextPage {
		return ProjectItem{}, fmt.Errorf("More then 100 project fields are not supported")
	}
	return m.AddProjectNextItem.ProjectNextItem, nil
}

func (c *Client) UpdateProjectItemField(projectID, itemID, fieldID githubv4.ID, value string, ctx context.Context) (ProjectItem, error) {
	var m struct {
		UpdateProjectNextItemField struct {
			ProjectNextItem struct {
				ID githubv4.ID
			}
		} `graphql:"updateProjectNextItemField(input: $input)"`
	}
	input := githubv4.UpdateProjectNextItemFieldInput{
		ProjectID: projectID,
		ItemID:    itemID,
		FieldID:   fieldID,
		Value:     githubv4.String(value),
	}

	err := c.client.Mutate(ctx, &m, input, nil)
	if err != nil {
		return ProjectItem{}, err
	}
	return ProjectItem{ID: m.UpdateProjectNextItemField.ProjectNextItem.ID}, nil
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

func (c *Client) ListOpenIssues(repo string, ctx context.Context) ([]Issue, error) {
	owner := strings.Split(repo, "/")[0]
	name := strings.Split(repo, "/")[1]

	var q struct {
		Repository struct {
			Issues struct {
				Nodes    []Issue
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"issues(first: 100, after: $issuesCursor, states: [OPEN])"` // 100 per page.
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}

	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(owner),
		"repositoryName":  githubv4.String(name),
		"issuesCursor":    (*githubv4.String)(nil), // Null after argument to get first page.
	}

	// Get Issues from all pages.
	var allIssues []Issue
	for {
		err := c.client.Query(ctx, &q, variables)
		if err != nil {
			return []Issue{}, err
		}
		allIssues = append(allIssues, q.Repository.Issues.Nodes...)
		if !q.Repository.Issues.PageInfo.HasNextPage {
			break
		}
		variables["issuesCursor"] = githubv4.NewString(q.Repository.Issues.PageInfo.EndCursor)
	}

	return allIssues, nil
}
