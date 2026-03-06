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
