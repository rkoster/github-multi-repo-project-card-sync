package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/rkoster/github-multi-repo-project-card-sync/config"
	"github.com/rkoster/github-multi-repo-project-card-sync/github"
	"github.com/shurcooL/githubv4"
)

func main() {
	logger := log.Default()

	var configFileName string
	flag.StringVar(&configFileName, "config", "config.yml", "config file name to load")

	flag.Parse()

	c, err := config.LoadConfig(configFileName)
	if err != nil {
		logger.Fatalf("failed to load config: %s", err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	var gh *github.Client
	if token != "" {
		gh = github.NewTokenClient(token, ctx)
	} else {
		private_key := strings.Replace(os.Getenv("GITHUB_PRIVATE_KEY"), `\n`, "\n", -1)
		app_id, err := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64)
		if err != nil {
			logger.Fatalf("failed to parse GITHUB_APP_ID: %s", err)
		}

		gh, err = github.NewAppClient(c.Project.Organization, app_id, private_key, ctx)
		if err != nil {
			logger.Fatalf("failed to setup app auth: %s", err)
		}
	}

	project, err := gh.GetOrganizationProject(c.Project.Organization, c.Project.Number, ctx)
	if err != nil {
		logger.Fatalf("failed to load github project: %s", err)
	}

	for i, repo := range c.Repositories {
		pullRequests, err := gh.ListOpenPullRequests(repo.Name, ctx)
		if err != nil {
			logger.Fatalf("failed to list pull requests for: %s got: %s", repo.Name, err)
		}
		issues, err := gh.ListOpenIssues(repo.Name, ctx)
		if err != nil {
			logger.Fatalf("failed to list issues for: %s got: %s", repo.Name, err)
		}
		fmt.Printf("\n[%d/%d] %s processing: %d ", i, len(c.Repositories), repo.Name, len(pullRequests)+len(issues))
		for _, pullRequest := range pullRequests {
			err = processPullRequest(pullRequest, project, repo, gh, ctx)
			if err != nil {
				logger.Fatalf("failed to process pull request: %s got: %s", pullRequest.URL, err)
			}
			fmt.Print(".")
		}
		for _, issue := range issues {
			err = processIssue(issue, project, repo, gh, ctx)
			if err != nil {
				logger.Fatalf("failed to process issue: %s got: %s", issue.URL, err)
			}
			fmt.Print(".")
		}
	}
}

func processIssue(issue github.Issue, project github.Project, repo config.Repository, gh *github.Client, ctx context.Context) error {
	item, err := gh.AddProjectItem(&project.ID, githubv4.NewID(issue.ID), ctx)
	if err != nil {
		return fmt.Errorf("failed to add project item got: %s", err)
	}

	for _, field := range repo.Fields {
		f, found := project.Fields.FindByName(field.Name)
		if !found {
			return fmt.Errorf("Project does not have field with name: %s", field.Name)
		}

		var input githubv4.ProjectV2FieldValue

		switch field.Type {
		case "draft":
			continue
		case "changes":
			continue
		case "author":
			input.Text = &issue.Author.Login
		case "last_activity":
			input.Date = &issue.TimelineItems.UpdatedAt
		case "default_single_select":
			currentValue, found := item.FieldValues.Nodes.FindByFieldName(field.Name)
			if found && currentValue.ProjectV2ItemFieldSingleSelectValue.OptionID != "" {
				continue
			}
			o, found := f.FindOptionByName(field.Value)
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, field.Value)
			}
			input.SingleSelectOptionID = &o.ID
		case "single_select":
			o, found := f.FindOptionByName(field.Value)
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, field.Value)
			}
			input.SingleSelectOptionID = &o.ID
		case "type":
			o, found := f.FindOptionByName("Issue")
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, "Issue")
			}
			input.SingleSelectOptionID = &o.ID
		default:
			return fmt.Errorf(
				"Only 'draft', 'author' and 'single_select' are currently supported values for field.type, given: %v", field.Type)
		}

		err := gh.UpdateProjectItemField(&project.ID, githubv4.NewID(item.ID), github.GetID(f), input, ctx)
		if err != nil {
			return fmt.Errorf("Failed to update Project Item Field got: %s", err)
		}
	}
	return nil
}

func processPullRequest(pullRequest github.PullRequest, project github.Project, repo config.Repository, gh *github.Client, ctx context.Context) error {
	item, err := gh.AddProjectItem(githubv4.NewID(project.ID), githubv4.NewID(pullRequest.ID), ctx)
	if err != nil {
		return fmt.Errorf("Failed to add Project Item got: %s", err)
	}

	for _, field := range repo.Fields {
		f, found := project.Fields.FindByName(field.Name)
		if !found {
			return fmt.Errorf("Project does not have field with name: %s", field.Name)
		}

		var input githubv4.ProjectV2FieldValue

		switch field.Type {
		case "draft":
			o, found := f.FindOptionByName(strconv.FormatBool(pullRequest.IsDraft))
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, field.Value)
			}
			input.SingleSelectOptionID = &o.ID
		case "changes":
			input.Number = pullRequest.Changes()
		case "author":
			input.Text = &pullRequest.Author.Login
		case "last_activity":
			input.Date = &pullRequest.TimelineItems.UpdatedAt
		case "default_single_select":
			currentValue, found := item.FieldValues.Nodes.FindByFieldName(field.Name)
			if found && currentValue.ProjectV2ItemFieldSingleSelectValue.OptionID != "" {
				continue
			}
			o, found := f.FindOptionByName(field.Value)
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, field.Value)
			}
			input.SingleSelectOptionID = &o.ID
		case "single_select":
			o, found := f.FindOptionByName(field.Value)
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, field.Value)
			}
			input.SingleSelectOptionID = &o.ID
		case "type":
			o, found := f.FindOptionByName("Pull Request")
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, "Pull Request")
			}
			input.SingleSelectOptionID = &o.ID
		default:
			return fmt.Errorf(
				"Only 'draft', 'author' and 'single_select' are currently supported values for field.type, given: %v", field.Type)
		}

		err := gh.UpdateProjectItemField(githubv4.NewID(project.ID),
			githubv4.NewID(item.ID), github.GetID(f), input, ctx)
		if err != nil {
			return fmt.Errorf("Failed to update Project Item Field got: %s", err)
		}
	}
	return nil
}
