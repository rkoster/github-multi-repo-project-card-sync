package github

import (
	"encoding/json"

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
		UpdatedAt githubv4.DateTime
	}
	Author struct {
		Login string
	}
}

type FileChange struct {
	Additions int
	Deletions int
	Path      string
}

func (pr PullRequest) Changes() int {
	var count int
	for _, change := range pr.Files.Nodes {
		if enry.IsVendor(change.Path) {
			continue
		}
		count += change.Additions + change.Deletions
	}
	return count
}

type Issue struct {
	ID            string
	URL           string
	TimelineItems struct {
		UpdatedAt githubv4.DateTime
	}
	Author struct {
		Login string
	}
}

type Project struct {
	ID     githubv4.ID
	Fields ProjectFields
}

type ProjectFields []ProjectField

type ProjectField struct {
	ID       githubv4.ID
	Name     string
	Settings string
}

type FieldOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

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

type ProjectItems []ProjectItem

type FieldValue struct {
	ProjectField ProjectField
	Value        string
}

type FieldValues []FieldValue

func (pf ProjectFields) FindByName(name string) (ProjectField, bool) {
	for _, field := range pf {
		if field.Name == name {
			return field, true
		}
	}
	return ProjectField{}, false
}

func (pf ProjectField) FindOptionByName(name string) (FieldOption, bool) {
	var settings struct {
		Options []FieldOption `json:"options"`
	}

	json.Unmarshal([]byte(pf.Settings), &settings)

	for _, option := range settings.Options {
		if option.Name == name {
			return option, true
		}
	}
	return FieldOption{}, false
}

func (fv FieldValues) FindByID(id githubv4.ID) (FieldValue, bool) {
	for _, field := range fv {
		if field.ProjectField.ID == id {
			return field, true
		}
	}
	return FieldValue{}, false
}
