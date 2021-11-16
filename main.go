package main

import (
	"context"
	"fmt"
	"os"

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
		for _, pullRequest := range pullRequests {
			item, err := gh.AddProjectItem(project.ID, pullRequest.ID, ctx)
			if err != nil {
				panic(err)
			}
			for _, field := range repo.Fields {
				f, found := project.Fields.FindByName(field.Name)
				if !found {
					panic(fmt.Errorf("Project does not have field with name: %s", field.Name))
				}

				o, found := f.FindOptionByName(field.Value)
				if !found {
					panic(fmt.Errorf("Project field: %s does not have an option: %s", field.Name, field.Value))
				}
				_, err := gh.UpdateProjectItemField(project.ID, item.ID, f.ID, o.ID, ctx)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
