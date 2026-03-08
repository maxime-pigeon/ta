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
- Never use `as`.
- Always run the commands below before finishing a task.

## Commands

```bash
bunx tsc --noEmit
bun test
```

## Architecture

`ta` is a TypeScript library (Bun) that parses JSON linter output and
posts inline pull-request review comments via the GitHub API.

**Data flow:** parsed JSON → `parse()` → `Lint[]` → `toComments()` →
`Comment[]` → `post()` → GitHub PR review

**Package layout:**

- `src/index.ts` — re-exports public API
- `src/review.ts` — `Lint` and `Comment` types, `toComments()`,
  `formatBody()`
- `src/parser.ts` — `parse()` and helpers; auto-detects ESLint,
  stylelint, and html-validate JSON formats
- `src/github.ts` — `post()` wraps GitHub REST API
  (`POST /repos/{owner}/{repo}/pulls/{pr}/reviews`)

**Format detection in `src/parser.ts`:** uses key presence to
distinguish formats — ESLint/html-validate use
`filePath`/`messages`/`ruleId`; stylelint uses
`source`/`warnings`/`rule`/`text`. Severity is normalized from either
number (ESLint: `2`=error) or string (stylelint: `"error"`).
