// ta reads JSON linter output from stdin and either prints findings to stdout
// (local mode) or posts them as inline GitHub review comments (CI mode).
//
// Local mode: any of --token, --repo, --pr, --sha is missing → print to stdout.
// CI mode: all four are present → post a GitHub pull-request review.
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/maximepigeon/ta/github"
	"github.com/maximepigeon/ta/review"
)

var cli struct {
	Token string `help:"GitHub token."                    env:"GITHUB_TOKEN"`
	Repo  string `help:"Repository in owner/repo format." env:"GITHUB_REPOSITORY"`
	PR    int    `help:"Pull request number."             env:"PR_NUMBER"`
	SHA   string `help:"Head commit SHA."                 env:"GITHUB_SHA"`
}

func main() {
	log.SetFlags(0)
	kong.Parse(&cli)

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("error reading stdin: %v", err)
	}

	lints, err := review.Parse(data)
	if err != nil {
		log.Fatalf("error parsing linter output: %v", err)
	}

	comments := review.ToComments(lints)

	if cli.Token == "" || cli.Repo == "" || cli.PR == 0 || cli.SHA == "" {
		for _, c := range comments {
			fmt.Printf("%s:%d: %s\n", c.Path, c.Line, c.Body)
		}
		return
	}

	if len(comments) == 0 {
		fmt.Println("no comments")
		return
	}

	if err := github.Post(cli.Token, cli.Repo, cli.PR, cli.SHA, comments); err != nil {
		log.Fatalf("error posting review: %v", err)
	}
	fmt.Printf("posted review with %d comment(s)\n", len(comments))
}
