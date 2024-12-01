package model

import "github.com/xanzy/go-gitlab"

type PipelineEvents struct {
	gitlab.PipelineEvent
	ErrorMsg string
	IsFailed bool
}
type ObjectAttributes struct {
	ID         int        `json:"id"`
	Ref        string     `json:"ref"`
	Tag        bool       `json:"tag"`
	Sha        string     `json:"sha"`
	BeforeSha  string     `json:"before_sha"`
	Status     string     `json:"status"`
	Stages     []string   `json:"stages"`
	CreatedAt  string     `json:"created_at"`
	FinishedAt string     `json:"finished_at"`
	Variables  []Variable `json:"variables"`
}

// variables in .gitlab.yml, not being used now
type Variable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type Project struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type User struct {
	Name     string `json:"name"`
	UserName string `json:"username"` //nolint:tagliatelle
}

type Commit struct {
	Message string `json:"message"`
}
