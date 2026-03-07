// Package github provides GitHub REST API helpers for posting inline review
// comments on pull requests.
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/maximepigeon/ta/review"
)

// reviewRequest is the payload for
// POST /repos/{owner}/{repo}/pulls/{pr}/reviews.
type reviewRequest struct {
	CommitID string           `json:"commit_id"`
	Body     string           `json:"body"`
	Event    string           `json:"event"`
	Comments []review.Comment `json:"comments"`
}

// buildReviewRequest converts comments into the GitHub review payload for
// the given commit SHA.
func buildReviewRequest(sha string, comments []review.Comment) reviewRequest {
	return reviewRequest{
		CommitID: sha,
		Body:     "linter remarks",
		Event:    "COMMENT",
		Comments: comments,
	}
}

// Post creates a GitHub pull-request review with one inline comment per
// finding via POST /repos/{owner}/{repo}/pulls/{pr}/reviews.
func Post(
	token, repo string, pr int, sha string, comments []review.Comment,
) error {
	payload, err := json.Marshal(buildReviewRequest(sha, comments))
	if err != nil {
		return fmt.Errorf("marshaling review: %w", err)
	}

	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/pulls/%d/reviews",
		repo, pr,
	)
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
