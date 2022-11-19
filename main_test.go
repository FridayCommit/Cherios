package main

import (
	"github.com/go-playground/webhooks/v6/github"
	githubApi "github.com/google/go-github/v48/github"
	"reflect"
	"testing"
)

func Test_handleCreateRepositoryEvent(t *testing.T) {
	type args struct {
		repositoryPayload github.RepositoryPayload
	}
	tt := github.RepositoryPayload{}
	tests := []struct {
		name string
		args args
	}{},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleCreateRepositoryEvent(tt.args.repositoryPayload)
		})
	}
}

func Test_initGitHubClient(t *testing.T) {
	tests := []struct {
		name string
		want *githubApi.Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := initGitHubClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initGitHubClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
