package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v43/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Client struct {
	client *githubv4.Client
}

func NewTokenClient(token string, ctx context.Context) *Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)

	return &Client{githubv4.NewClient(httpClient)}
}

func NewAppClient(org string, app_id int64, private_key string, ctx context.Context) (*Client, error) {
	appItr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, app_id, []byte(private_key))
	if err != nil {
		return nil, err
	}
	client := &http.Client{Transport: appItr}
	gh := github.NewClient(client)
	installation, _, err := gh.Apps.FindOrganizationInstallation(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to find installation for organization: %s got: %s", org, err)
	}

	itr, err := ghinstallation.New(http.DefaultTransport, app_id, *installation.ID, []byte(private_key))
	if err != nil {
		return nil, err
	}

	return &Client{githubv4.NewClient(&http.Client{Transport: itr})}, nil
}

func (c *Client) GetOrganizationProject(org string, projectNumber int, ctx context.Context) (Project, error) {
	var q struct {
		Organization struct {
			ProjectV2 struct {
				ID     githubv4.ID
				Fields struct {
					Nodes    ProjectFields
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"fields(first: 100, after: $fieldsCursor)"`
			} `graphql:"projectV2(number: $projectNumber)"`
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
		project.ID = &q.Organization.ProjectV2.ID
		project.Fields = append(project.Fields, q.Organization.ProjectV2.Fields.Nodes...)
		if !q.Organization.ProjectV2.Fields.PageInfo.HasNextPage {
			break
		}
		variables["fieldsCursor"] = githubv4.NewString(q.Organization.ProjectV2.Fields.PageInfo.EndCursor)
	}

	return project, nil
}

func (c *Client) AddProjectItem(projectID, itemID *githubv4.ID, ctx context.Context) (ProjectItem, error) {
	var m struct {
		AddProjectV2ItemById struct {
			Item ProjectItem
		} `graphql:"addProjectV2ItemById(input: $input)"`
	}
	input := githubv4.AddProjectV2ItemByIdInput{
		ProjectID: projectID,
		ContentID: itemID,
	}

	err := c.client.Mutate(ctx, &m, input, nil)
	if err != nil {
		return ProjectItem{}, err
	}

	if m.AddProjectV2ItemById.Item.FieldValues.PageInfo.HasNextPage {
		return ProjectItem{}, fmt.Errorf("More then 100 project fields are not supported")
	}
	return m.AddProjectV2ItemById.Item, nil
}

type FieldInput struct {
	Date                 githubv4.Date
	Number               githubv4.Int
	SingleSelectOptionId githubv4.ID
	Text                 githubv4.String
}

func (c *Client) UpdateProjectItemField(projectID, itemID, fieldID githubv4.ID, input githubv4.ProjectV2FieldValue, ctx context.Context) error {
	var m struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationId string
		} `graphql:" updateProjectV2ItemFieldValue(input: $input)"`
	}
	i := githubv4.UpdateProjectV2ItemFieldValueInput{
		ProjectID: projectID,
		ItemID:    itemID,
		FieldID:   &fieldID,
		Value:     input,
	}

	return c.client.Mutate(ctx, &m, i, nil)
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
