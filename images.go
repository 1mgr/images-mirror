package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/distribution/reference"
)

const (
	legacyDefaultDomain = "index.docker.io"
	defaultDomain       = "docker.io"
	otherDomain         = "registry-1.docker.io"
	officialRepoName    = "library"
)

// splitDockerImageParts splits a repository name to domain and remotename string.
// If no valid domain is found, the default domain is used. Repository name
// needs to be already validated before.
func splitDockerImageParts(name string) (domain, remainder, tag string) {
	i := strings.IndexRune(name, '/')
	if i == -1 || (!strings.ContainsAny(name[:i], ".:") && name[:i] != "localhost") {
		domain, remainder = defaultDomain, name
	} else {
		domain, remainder = name[:i], name[i+1:]
	}
	if domain == legacyDefaultDomain {
		domain = defaultDomain
	}
	if domain == otherDomain {
		domain = defaultDomain
	}
	if domain == defaultDomain && !strings.ContainsRune(remainder, '/') {
		remainder = officialRepoName + "/" + remainder
	}
	tag = "latest"
	if i := strings.IndexRune(remainder, ':'); i != -1 {
		remainder, tag = remainder[:i], remainder[i+1:]
	}
	return
}

func isValidImage(image string) bool {
	matches := reference.ReferenceRegexp.FindStringSubmatch(image)
	if matches == nil || len(matches) == 0 {
		return false
	}
	imagePart, _, _ := matches[1], matches[2], matches[3]
	domain, _, _ := splitDockerImageParts(imagePart)

	if domain != "" && domain != defaultDomain {
		return false
	}

	return true
}

func shortenRemainder(remainder string) string {
	if strings.HasPrefix(remainder, officialRepoName+"/") {
		return strings.TrimPrefix(remainder, officialRepoName+"/")
	}
	return remainder
}

func imageExists(image string) bool {
	client := &http.Client{Timeout: time.Duration(10) * time.Second}
	_, remainder, tag := splitDockerImageParts(image)
	resp, err := client.Get(fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/%s", remainder, tag))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	return true
}
