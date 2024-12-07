package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type StatusWriter interface {
	Write(status string)
}

type GithubClientConfig struct {
	Token         string
	OrgRepo       string
	Timeout       time.Duration
	CheckInterval time.Duration
}

type GitHubClient struct {
	Config GithubClientConfig
}

func NewGitHubClient(sw StatusWriter, config GithubClientConfig) *GitHubClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.CheckInterval == 0 {
		config.CheckInterval = 1 * time.Second
	}
	return &GitHubClient{Config: config}
}

func (g *GitHubClient) callAPI(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.Config.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: g.Config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}

	return resp, nil
}

func (g *GitHubClient) LaunchGithubAction(image, id string) error {
	_, remainder, tag := splitDockerImageParts(image)
	remainder = shortenRemainder(remainder)

	payload := map[string]interface{}{
		"ref": "main",
		"inputs": map[string]string{
			"image_name": fmt.Sprintf("%s:%s", remainder, tag),
			"id":         id,
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshall json: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/workflows/mirror.yml/dispatches", g.Config.OrgRepo)

	_, err = g.callAPI("POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to call dispatch workflow: %w", err)
	}

	return nil
}

type runsResponse struct {
	WorkflowRuns []struct {
		ID      int    `json:"id"`
		Status  string `json:"status"`
		JobsURL string `json:"jobs_url"`
	} `json:"workflow_runs"`
}

func (g *GitHubClient) getLastRuns() (*runsResponse, error) {
	var dateFilter string = "%3E" + time.Now().UTC().Add(-5*time.Minute).Format("2006-01-02T15:04")
	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/runs?created=%s", g.Config.OrgRepo, dateFilter)
	resp, err := g.callAPI("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("call api ended with status: %s", resp.Status)
	}

	var runs runsResponse
	if err := json.NewDecoder(resp.Body).Decode(&runs); err != nil {
		return nil, fmt.Errorf("failed to decode response json: %w", err)
	}

	return &runs, nil
}

type jobsResponse struct {
	Jobs []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		RunID int    `json:"run_id"`
		Steps []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"steps"`
	} `json:"jobs"`
}

func (c *GitHubClient) getJobs(url string) (*jobsResponse, error) {
	resp, err := c.callAPI("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("call API ended with status: %s", resp.Status)
	}

	var jobs jobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		return nil, fmt.Errorf("failed to decode response json: %w", err)
	}

	return &jobs, nil
}

func (g *GitHubClient) FollowWorkflowRun(sw StatusWriter, id string) error {

	var workflowID = ""
	var workflow_jobs_url = ""
	for workflowID == "" {
		// Get the list of workflow runs
		runs, err := g.getLastRuns()
		if err != nil {
			return fmt.Errorf("failed to get last runs: %w", err)
		}

		if len(runs.WorkflowRuns) == 0 {
			sw.Write("⏳ No workflow runs found yet")
			time.Sleep(g.Config.CheckInterval)
			continue // Retry
		}

		// Get the latest workflow run
		for _, run := range runs.WorkflowRuns {
			jobs, err := g.getJobs(run.JobsURL)
			if err != nil {
				return fmt.Errorf("failed to get jobs: %w", err)
			}
			for _, job := range jobs.Jobs {
				for _, step := range job.Steps {
					if step.Name == id {
						workflowID = fmt.Sprintf("%d", job.RunID)
						break
					}
				}
				if workflowID != "" {
					workflow_jobs_url = run.JobsURL
					break
				}
			}
		}
	}

	sw.Write(fmt.Sprintf("✨ Workflow run found – https://github.com/%s/actions/runs/%s", g.Config.OrgRepo, workflowID))
	stepsStatus := make(map[string]string)

	for {
		jobs, err := g.getJobs(workflow_jobs_url)
		if err != nil {
			return fmt.Errorf("failed to get jobs: %w", err)
		}

		allFinished := true
		for _, job := range jobs.Jobs {
			if job.Name == "Workflow ID Provider" {
				continue
			}
			for i, step := range job.Steps {

				if step.Status != "completed" {
					allFinished = false
				} else {
					if _, exists := stepsStatus[step.Name]; !exists {
						stepsStatus[step.Name] = step.Status
						sw.Write(fmt.Sprintf("🔄 Completed step %d/%d: %s", i+1, len(job.Steps), step.Name))
					}
				}
			}
		}

		if allFinished {
			sw.Write("🎉 All workflow steps completed")
			break
		}

		time.Sleep(g.Config.CheckInterval)
	}

	return nil
}
