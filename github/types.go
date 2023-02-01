package github

import (
	enry "github.com/go-enry/go-enry/v2"
	"github.com/shurcooL/githubv4"
)

type PullRequest struct {
	ID      string
	URL     string
	IsDraft bool
	Files   struct {
		Nodes []FileChange
	} `graphql:"files(first: 100)"`
	TimelineItems struct {
		UpdatedAt githubv4.Date
	}
	Author struct {
		Login githubv4.String
	}
}

type FileChange struct {
	Additions int
	Deletions int
	Path      string
}

func (pr PullRequest) Changes() *githubv4.Float {
	var count githubv4.Float
	for _, change := range pr.Files.Nodes {
		if enry.IsVendor(change.Path) {
			continue
		}
		count += githubv4.Float(change.Additions) + githubv4.Float(change.Deletions)
	}
	return &count
}

type Issue struct {
	ID            githubv4.ID
	URL           string
	TimelineItems struct {
		UpdatedAt githubv4.Date
	}
	Author struct {
		Login githubv4.String
	}
}

type Project struct {
	ID     githubv4.ID
	Fields ProjectFields
}

type ProjectItems []ProjectItem

type ProjectItem struct {
	ID          githubv4.ID
	FieldValues struct {
		Nodes    FieldValues
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage bool
		}
	} `graphql:"fieldValues(first: 100)"`
}
