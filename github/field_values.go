package github

import "github.com/shurcooL/githubv4"

type FieldValues []FieldValue

type FieldValue struct {
	ProjectV2ItemFieldDateValue         `graphql:"... on ProjectV2ItemFieldDateValue"`
	ProjectV2ItemFieldNumberValue       `graphql:"... on ProjectV2ItemFieldNumberValue"`
	ProjectV2ItemFieldTextValue         `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
	ProjectV2ItemFieldSingleSelectValue `graphql:"... on ProjectV2ItemFieldTextValue"`
}

type ProjectV2ItemFieldDateValue struct {
	ID    githubv4.ID
	Date  githubv4.String
	Field ProjectField
}

type ProjectV2ItemFieldNumberValue struct {
	ID     githubv4.ID
	Number githubv4.Float
	Field  ProjectField
}

type ProjectV2ItemFieldTextValue struct {
	ID       githubv4.ID
	OptionID githubv4.String
	Field    ProjectField
}

type ProjectV2ItemFieldSingleSelectValue struct {
	ID    githubv4.ID
	Text  githubv4.String
	Field ProjectField
}

func (fv FieldValues) FindByID(id githubv4.ID) (FieldValue, bool) {
	for _, field := range fv {
		if field.ProjectV2ItemFieldDateValue.ID == id ||
			field.ProjectV2ItemFieldNumberValue.ID == id ||
			field.ProjectV2ItemFieldTextValue.ID == id ||
			field.ProjectV2ItemFieldSingleSelectValue.ID == id {
			return field, true
		}
	}
	return FieldValue{}, false
}
