package github

type PullRequest struct {
	ID      string
	URL     string
	IsDraft bool
	Author  struct {
		Login string
	}
}

type Project struct {
	ID     string
	Fields []ProjectField
}

type ProjectField struct {
	ID       string
	Name     string
	Settings interface{}
}
