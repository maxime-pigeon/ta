# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working
with code in this repository.

## Rules

- Prefer simple and idiomatic code.
- Keep lines at most 80 characters.
- Never ignore errors, always handle them.
- Always add context to errors.
- Always document code.
- Avoid dependencies when possible.
- Always run the commands below before finishing a task.

## Commands

```bash
go fmt ./...
go vet ./...
go tool staticcheck ./...
go test ./...
```

## Architecture

`ta` is a GitHub Action (defined in `action.yml`) that reads JSON linter
output from stdin and posts inline pull-request review comments via the
GitHub API.

**Data flow:** stdin JSON → `review.Parse()` → `[]Lint` →
`review.ToComments()` → `[]Comment` → `github.Post()` → GitHub PR review

**Two operating modes** (determined at runtime in `main.go`):

- **Local mode**: any of `--token`, `--repo`, `--pr`, `--sha` is missing
  → prints `path:line: message` to stdout
- **CI mode**: all four present → calls `github.Post()` to create a
  GitHub pull-request review

**Package layout:**

- `main.go` — CLI parsing (via `kong`), stdin reading, mode dispatch
- `review/review.go` — `Lint` and `Comment` types, `ToComment()`,
  `ToComments()`, `formatBody()`
- `review/parser.go` — `Parse()` and helpers; auto-detects ESLint,
  stylelint, and html-validate JSON formats
- `github/github.go` — `Post()` wraps GitHub REST API
  (`POST /repos/{owner}/{repo}/pulls/{pr}/reviews`)

**Format detection in `review/parser.go`:** uses key presence to
distinguish formats — ESLint/html-validate use
`filePath`/`messages`/`ruleId`; stylelint uses
`source`/`warnings`/`rule`/`text`. Severity is normalized from either
integer (ESLint: `2`=error) or string (stylelint: `"error"`).
