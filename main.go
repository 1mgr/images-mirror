package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

type Server struct {
	gh *GitHubClient
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	r := chi.NewRouter()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Println("GITHUB_TOKEN is required")
		os.Exit(1)
	}

	server := Server{
		gh: NewGitHubClient(nil, GithubClientConfig{
			Token:         token,
			OrgRepo:       "1mgr/images-mirror",
			Timeout:       30 * time.Second,
			CheckInterval: 1 * time.Second,
		}),
	}

	r.Get("/*", server.MirrorImageHandler)

	http.ListenAndServe(":8080", r)
}

func (server *Server) MirrorImageHandler(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Path
	sourceIP := req.Header.Get("CF-Connecting-IP")
	if sourceIP == "" {
		sourceIP = req.RemoteAddr
	}
	log.Printf("Received request: query=%s, sourceIP=%s\n", query, sourceIP)

	userAgent := strings.ToLower(req.Header.Get("User-Agent"))
	if !strings.Contains(userAgent, "curl") && !strings.Contains(userAgent, "wget") {
		http.Redirect(w, req, "https://github.com/1mgr/image-mirrors", http.StatusFound)
		return
	}

	image := strings.TrimPrefix(req.URL.Path, "/")
	if image == "" {
		httpError(w, 400, "image not passed")
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	writeLine(w, fmt.Sprintf("⬇️ Received request to mirror image: %s", image))
	if !isValidImage(image) {
		httpError(w, http.StatusBadRequest, fmt.Sprintf("❌ Invalid image: %s. Image name must be valid docker hub image with version tag. For example: postgres:16", image))
		return
	}
	if !imageExists(image) {
		httpError(w, http.StatusBadRequest, fmt.Sprintf("❌ Image not found in dockerhub: %s", image))
		return
	}
	id := randID()
	err := server.gh.LaunchGithubAction(image, id)
	if err != nil {
		httpError(w, http.StatusInternalServerError, fmt.Sprintf("❌ Failed to launch GitHub action: %v", err))
		return
	}
	writeLine(w, "🚀 Launched github action workflow")
	if err := server.gh.FollowWorkflowRun(makeStatusWriter(w), id); err != nil {
		httpError(w, http.StatusInternalServerError, fmt.Sprintf("❌ Failed to follow workflow run: %v", err))
		return
	}

	writeLine(w, "ℹ️ Pull your image from the mirror:")
	writeLine(w, "docker pull ghcr.io/1mgr/"+image)
}

func randID() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		b[i] = letters[num.Int64()]
	}
	return string(b)
}
