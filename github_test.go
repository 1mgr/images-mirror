package main

import (
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestGitHubClient_IsImageAlreadyMirrored(t *testing.T) {
	godotenv.Load()
	cfg := GithubClientConfig{
		Token:         os.Getenv("GITHUB_TOKEN"),
		Org:           "1mgr",
		Timeout:       30 * time.Second,
		CheckInterval: 1 * time.Second,
	}
	type args struct {
		image string
	}
	tests := []struct {
		name                   string
		args                   args
		mirrored               bool
		sinceLastMirrorChecker func(time.Duration) bool
	}{
		{
			name:                   "postgres 11",
			args:                   args{image: "postgres:11"},
			mirrored:               true,
			sinceLastMirrorChecker: func(d time.Duration) bool { return d > time.Hour*24 },
		},
		{
			name:                   "postgres 16",
			args:                   args{image: "postgres:16"},
			mirrored:               true,
			sinceLastMirrorChecker: func(d time.Duration) bool { return d < time.Hour*12 && d > time.Hour*1 },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GitHubClient{
				Config: cfg,
			}
			got, got1 := g.IsImageAlreadyMirrored(tt.args.image)
			if got != tt.mirrored {
				t.Errorf("GitHubClient.IsImageAlreadyMirrored() got = %v, want %v", got, tt.mirrored)
			}
			if !tt.sinceLastMirrorChecker(got1) {
				t.Errorf("GitHubClient.IsImageAlreadyMirrored() got1 = %v", got1)
			}
		})
	}
}
