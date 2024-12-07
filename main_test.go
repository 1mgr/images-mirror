package main

import (
	"testing"
)

func Test_splitDockerDomain(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name          string
		args          args
		wantDomain    string
		wantRemainder string
		wantTag       string
	}{
		{
			name:          "default docker hub",
			args:          args{name: "docker.io/library/alpine"},
			wantDomain:    "docker.io",
			wantRemainder: "library/alpine",
			wantTag:       "latest",
		},
		{
			name:          "short image",
			args:          args{name: "alpine"},
			wantDomain:    "docker.io",
			wantRemainder: "library/alpine",
			wantTag:       "latest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDomain, gotRemainder, tag := splitDockerImageParts(tt.args.name)
			if gotDomain != tt.wantDomain {
				t.Errorf("splitDockerDomain() gotDomain = %v, want %v", gotDomain, tt.wantDomain)
			}
			if gotRemainder != tt.wantRemainder {
				t.Errorf("splitDockerDomain() gotRemainder = %v, want %v", gotRemainder, tt.wantRemainder)
			}
			if tag != tt.wantTag {
				t.Errorf("splitDockerDomain() tag = %v, want %v", tag, tt.wantTag)
			}
		})
	}
}

func Test_isValidImage(t *testing.T) {
	type args struct {
		image string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid dockerhub image",
			args: args{image: "docker.io/library/alpine"},
			want: true,
		},
		{
			name: "valid dockerhub image with latest tag",
			args: args{image: "docker.io/library/alpine:latest"},
			want: true,
		},
		{
			name: "invalid ghcr image",
			args: args{image: "ghcr.io/library/alpine"},
			want: false,
		},
		{
			name: "invalid quay image",
			args: args{image: "quay.io/library/alpine"},
			want: false,
		},
		{
			name: "valid short image",
			args: args{image: "alpine"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidImage(tt.args.image); got != tt.want {
				t.Errorf("isValidImage(%s) = %v, want %v", tt.args.image, got, tt.want)
			}
		})
	}
}

func Test_imageExists(t *testing.T) {
	type args struct {
		image string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "full postgres image",
			args: args{image: "docker.io/library/postgres:16"},
			want: true,
		},
		{
			name: "short postgres image",
			args: args{image: "postgres:16"},
			want: true,
		},
		{
			name: "very short image",
			args: args{image: "alpine"},
			want: true,
		},
		{
			name: "invalid image",
			args: args{image: "notexist/postgres:16"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := imageExists(tt.args.image); got != tt.want {
				t.Errorf("imageExists(%s) = %v, want %v", tt.args.image, got, tt.want)
			}
		})
	}
}
