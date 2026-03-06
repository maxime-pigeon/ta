// ta is a teaching-assistant tool that runs configured linters against changed
// files in a pull request and posts the comments as inline GitHub review
// comments.
//
// Local mode: pass files as arguments — comments are printed to stdout.
// CI mode: omit file arguments — changed files are fetched from the PR and
// comments are posted as a GitHub review.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

"github.com/BurntSushi/toml"
	"github.com/alecthomas/kong"
	"github.com/maximepigeon/ta/github"
	"github.com/maximepigeon/ta/linter"
)

type tomlConfig struct {
	Linter []linter.Linter `toml:"linter"`
}

// cli holds all command-line arguments parsed by Kong.
var cli struct {
	Token  string   `help:"GitHub token."                        env:"GITHUB_TOKEN"`
	Repo   string   `help:"Repository in owner/repo format."     env:"GITHUB_REPOSITORY"`
	PR     int      `help:"Pull request number."                 env:"PR_NUMBER"`
	SHA    string   `help:"Head commit SHA."                     env:"GITHUB_SHA"`
	Config string   `help:"Path to TOML config file."            default:"ta.toml"`
	Files  []string `arg:"" optional:"" help:"Files to lint (local mode)."`
}

// runLocal lints the provided files and prints comments to stdout.
func runLocal(linters []linter.Linter, files []string, dir string) {
	lints, err := linter.RunAll(linters, files, dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range linter.ToComments(lints) {
		fmt.Printf("%s:%d: %s\n", c.Path, c.Line, c.Body)
	}
}

// absolutePaths prepends dir to each file path. The GitHub API returns paths
// relative to the repository root (e.g. "src/main.js"), but linters require
// absolute paths to locate files on disk. In CI, dir is GITHUB_WORKSPACE
// (e.g. "/github/workspace"), which is where Actions checks out the repo.
func absolutePaths(files []string, dir string) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = filepath.Join(dir, f)
	}
	return out
}

// runCI fetches changed files from the PR and posts comments as a GitHub review.
func runCI(linters []linter.Linter) {
	if cli.PR == 0 {
		log.Fatal("error: --pr must be a valid PR number (or pass files as arguments)")
	}
	if cli.Token == "" {
		log.Fatal("error: --token is required")
	}
	if cli.Repo == "" {
		log.Fatal("error: --repo is required")
	}
	files, err := github.GetChangedFiles(cli.Token, cli.Repo, cli.PR)
	if err != nil {
		log.Fatalf("error getting changed files: %v", err)
	}
	if len(files) == 0 {
		fmt.Println("no files changed")
		return
	}
	workspace := os.Getenv("GITHUB_WORKSPACE")
	lints, err := linter.RunAll(linters, absolutePaths(files, workspace), filepath.Dir(cli.Config))
	if err != nil {
		log.Fatal(err)
	}
	comments := linter.ToComments(lints)
	for i, c := range comments {
		comments[i].Path = strings.TrimPrefix(c.Path, workspace+"/")
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

func main() {
	log.SetFlags(0) // suppress default timestamp prefix on error output
	kong.Parse(&cli)

	var cfg tomlConfig
	if _, err := toml.DecodeFile(cli.Config, &cfg); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	if len(cli.Files) > 0 {
		runLocal(cfg.Linter, cli.Files, filepath.Dir(cli.Config))
	} else {
		runCI(cfg.Linter)
	}
}
