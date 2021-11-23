package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/rkoster/github-multi-repo-project-card-sync/config"
	"github.com/rkoster/github-multi-repo-project-card-sync/github"
)

func main() {
	c, err := config.LoadConfig("config.yml")
	if err != nil {
		panic(err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	gh := github.NewClient(token, ctx)

	project, err := gh.GetOrganizationProject(c.Project.Organization, c.Project.Number, ctx)
	if err != nil {
		panic(err)
	}

	for _, repo := range c.Repositories {
		pullRequests, err := gh.ListOpenPullRequests(repo.Name, ctx)
		if err != nil {
			panic(err)
		}
		err = processPullRequests(pullRequests, project, repo, gh, ctx)
		if err != nil {
			panic(err)
		}
		issues, err := gh.ListOpenIssues(repo.Name, ctx)
		if err != nil {
			panic(err)
		}
		err = processIssues(issues, project, repo, gh, ctx)
		if err != nil {
			panic(err)
		}
	}
}

func processIssues(issues []github.Issue, project github.Project, repo config.Repository, gh *github.Client, ctx context.Context) error {
	for _, issue := range issues {
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
	}
	return nil
}

func processPullRequests(pullRequests []github.PullRequest, project github.Project, repo config.Repository, gh *github.Client, ctx context.Context) error {
	for _, pullRequest := range pullRequests {
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
	}
	return nil
}
