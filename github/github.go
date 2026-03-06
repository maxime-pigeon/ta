// Package github provides GitHub REST API helpers for fetching pull-request
// files and posting inline review comments.
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/maximepigeon/ta/linter"
)

// prFile is the subset of a GitHub pull-request file object that we need.
type prFile struct {
	Filename string `json:"filename"`
}

// reviewRequest is the payload for POST /repos/{owner}/{repo}/pulls/{pr}/reviews.
type reviewRequest struct {
	CommitID string           `json:"commit_id"`
	Body     string           `json:"body"`
	Event    string           `json:"event"`
	Comments []linter.Comment `json:"comments"`
}

// buildReviewRequest converts comments into the GitHub review payload for
// the given commit SHA.
func buildReviewRequest(sha string, comments []linter.Comment) reviewRequest {
	return reviewRequest{
		CommitID: sha,
		Body:     "linter remarks",
		Event:    "COMMENT",
		Comments: comments,
	}
}

// GetChangedFiles returns the paths of all files changed in the given pull
// request by calling the GitHub REST API.
func GetChangedFiles(token, repo string, pr int) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls/%d/files", repo, pr)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching PR files: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, body)
	}

	var files []prFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("decoding PR files response: %w", err)
	}

	filenames := make([]string, len(files))
	for i, f := range files {
		filenames[i] = f.Filename
	}
	return filenames, nil
}

// Post creates a GitHub pull-request review with one inline comment per
// finding via POST /repos/{owner}/{repo}/pulls/{pr}/reviews.
func Post(token, repo string, pr int, sha string, comments []linter.Comment) error {
	payload, err := json.Marshal(buildReviewRequest(sha, comments))
	if err != nil {
		return fmt.Errorf("marshaling review: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls/%d/reviews", repo, pr)
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("posting review: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, body)
	}
	return nil
}
