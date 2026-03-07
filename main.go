// ta reads JSON linter output from stdin and either prints findings to
// stdout or posts them as inline GitHub review comments.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/maximepigeon/ta/github"
	"github.com/maximepigeon/ta/review"
)

// stripWorkspacePrefix removes the GITHUB_WORKSPACE prefix from each
// comment's path so paths are relative to the repository root, as required
// by the GitHub PR review API.
func stripWorkspacePrefix(comments []review.Comment, workspace string) {
	if workspace == "" {
		return
	}
	prefix := strings.TrimSuffix(workspace, "/") + "/"
	for i := range comments {
		comments[i].Path = strings.TrimPrefix(comments[i].Path, prefix)
	}
}

func main() {
	log.SetFlags(0)

	mode := flag.String("mode", "stdout",
		`output mode: "stdout" prints findings to stdout; `+
			`"ci" posts inline GitHub review comments`)
	flag.Parse()

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("error reading stdin: %v", err)
	}

	lints, err := review.Parse(data)
	if err != nil {
		log.Fatalf("error parsing linter output: %v", err)
	}

	comments := review.ToComments(lints)

	switch *mode {
	case "stdout":
		for _, c := range comments {
			fmt.Printf("%s:%d: %s\n", c.Path, c.Line, c.Body)
		}
	case "ci":
		token := os.Getenv("GITHUB_TOKEN")
		repo := os.Getenv("GITHUB_REPOSITORY")
		sha := os.Getenv("GITHUB_SHA")
		pr, err := strconv.Atoi(os.Getenv("PR_NUMBER"))
		if err != nil {
			log.Fatalf("error parsing PR_NUMBER: %v", err)
		}

		if token == "" || repo == "" || pr == 0 || sha == "" {
			log.Fatal("CI mode requires GITHUB_TOKEN, GITHUB_REPOSITORY, " +
				"PR_NUMBER, and GITHUB_SHA")
		}

		stripWorkspacePrefix(comments, os.Getenv("GITHUB_WORKSPACE"))
		workspace := os.Getenv("GITHUB_WORKSPACE")
		stripWorkspacePrefix(comments, workspace)

		if len(comments) == 0 {
			log.Println("no comments")
			return
		}

		if err := github.Post(token, repo, pr, sha, comments); err != nil {
			log.Fatalf("error posting review: %v", err)
		}
		fmt.Printf("posted review with %d comment(s)\n", len(comments))
	}
}
