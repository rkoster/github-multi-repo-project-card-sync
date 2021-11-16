module github.com/rkoster/github-multi-repo-project-card-sync

go 1.16

require (
	github.com/shurcooL/githubv4 v0.0.0-20210922025249-6831e00d857f
	github.com/shurcooL/graphql v0.0.0-20200928012149-18c5c3165e3a // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/shurcooL/githubv4 => github.com/rkoster/githubv4 v0.0.0-20211116152855-3c0e9ae996cd
