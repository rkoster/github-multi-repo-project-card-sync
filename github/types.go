package github

import (
	"github.com/shurcooL/githubv4"
)

type PullRequest struct {
	ID      string
	URL     string
	IsDraft bool
	Author  struct {
		Login string
	}
}

type Project struct {
	ID     githubv4.ID
	Fields []ProjectField
}

type ProjectField struct {
	ID       githubv4.ID
	Name     string
	Settings interface{}
}

type ProjectItem struct {
	ID githubv4.ID
}

type addProjectNextItem struct {
	ProjectID githubv4.ID `json:"projectId"`
	ContentID githubv4.ID `json:"contentId"`
}
