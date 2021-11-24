package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/rkoster/github-multi-repo-project-card-sync/config"
	"github.com/rkoster/github-multi-repo-project-card-sync/github"
)

func main() {
	logger := log.Default()

	c, err := config.LoadConfig("config.yml")
	if err != nil {
		logger.Fatalf("failed to load config: %s", err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	gh := github.NewClient(token, ctx)

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
	item, err := gh.AddProjectItem(project.ID, issue.ID, ctx)
	if err != nil {
		return err
	}

	for _, field := range repo.Fields {
		f, found := project.Fields.FindByName(field.Name)
		if !found {
			return fmt.Errorf("Project does not have field with name: %s", field.Name)
		}

		var value string

		switch field.Type {
		case "draft":
			continue
		case "author":
			value = issue.Author.Login
		case "last_activity":
			value = issue.TimelineItems.UpdatedAt.String()
		case "default_single_select":
			currentValue, found := item.FieldValues.Nodes.FindByID(f.ID)
			if found && currentValue.Value != "" {
				continue
			}
			o, found := f.FindOptionByName(field.Value)
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, field.Value)
			}
			value = o.ID
		case "single_select":
			o, found := f.FindOptionByName(field.Value)
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, field.Value)
			}
			value = o.ID
		case "type":
			o, found := f.FindOptionByName("Issue")
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, "Issue")
			}
			value = o.ID
		default:
			return fmt.Errorf(
				"Only 'draft', 'author' and 'single_select' are currently supported values for field.type, given: %v", field.Type)
		}

		_, err := gh.UpdateProjectItemField(project.ID, item.ID, f.ID, value, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func processPullRequest(pullRequest github.PullRequest, project github.Project, repo config.Repository, gh *github.Client, ctx context.Context) error {
	item, err := gh.AddProjectItem(project.ID, pullRequest.ID, ctx)
	if err != nil {
		return err
	}

	for _, field := range repo.Fields {
		f, found := project.Fields.FindByName(field.Name)
		if !found {
			return fmt.Errorf("Project does not have field with name: %s", field.Name)
		}

		var value string

		switch field.Type {
		case "draft":
			o, found := f.FindOptionByName(strconv.FormatBool(pullRequest.IsDraft))
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, field.Value)
			}
			value = o.ID
		case "author":
			value = pullRequest.Author.Login
		case "last_activity":
			value = pullRequest.TimelineItems.UpdatedAt.String()
		case "default_single_select":
			currentValue, found := item.FieldValues.Nodes.FindByID(f.ID)
			if found && currentValue.Value != "" {
				continue
			}
			o, found := f.FindOptionByName(field.Value)
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, field.Value)
			}
			value = o.ID
		case "single_select":
			o, found := f.FindOptionByName(field.Value)
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, field.Value)
			}
			value = o.ID
		case "type":
			o, found := f.FindOptionByName("Pull Request")
			if !found {
				return fmt.Errorf("Project field: %s does not have an option: %s", field.Name, "Issue")
			}
			value = o.ID
		default:
			return fmt.Errorf(
				"Only 'draft', 'author' and 'single_select' are currently supported values for field.type, given: %v", field.Type)
		}

		_, err := gh.UpdateProjectItemField(project.ID, item.ID, f.ID, value, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
