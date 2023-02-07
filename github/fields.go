package github

import "github.com/shurcooL/githubv4"

type ProjectFields []ProjectField

type ProjectField struct {
	ProjectV2Field             `graphql:"... on ProjectV2Field"`
	ProjectV2IterationField    `graphql:"... on ProjectV2IterationField"`
	ProjectV2SingleSelectField `graphql:"... on ProjectV2SingleSelectField"`
}

type ProjectV2Field struct {
	ID   githubv4.ID
	Name githubv4.String
}

type ProjectV2IterationField = ProjectV2Field

type ProjectV2SingleSelectField struct {
	ID      githubv4.ID
	Name    githubv4.String
	Options []FieldOption
}

type FieldOption struct {
	ID   githubv4.String
	Name githubv4.String
}

func (pf ProjectFields) FindByName(name string) (ProjectField, bool) {
	n := githubv4.String(name)
	for _, field := range pf {
		if field.ProjectV2Field.Name == n || field.ProjectV2IterationField.Name == n || field.ProjectV2SingleSelectField.Name == n {
			return field, true
		}
	}
	return ProjectField{}, false
}

func (pf ProjectField) FindOptionByName(name string) (FieldOption, bool) {
	for _, option := range pf.ProjectV2SingleSelectField.Options {
		if option.Name == githubv4.String(name) {
			return option, true
		}
	}
	return FieldOption{}, false
}

func GetID(pf ProjectField) githubv4.ID {
	if pf.ProjectV2Field.ID != nil {
		return pf.ProjectV2Field.ID
	}
	if pf.ProjectV2IterationField.ID != nil {
		return pf.ProjectV2Field.ID
	}
	return pf.ProjectV2SingleSelectField.ID
}

func GetName(pf ProjectField) string {
	if pf.ProjectV2Field.Name != "" {
		return string(pf.ProjectV2Field.Name)
	}
	if pf.ProjectV2IterationField.Name != "" {
		return string(pf.ProjectV2Field.Name)
	}
	return string(pf.ProjectV2SingleSelectField.Name)
}
