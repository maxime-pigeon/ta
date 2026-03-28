- Prioritize code correctness and clarity. Speed and efficiency are
  secondary priorities unless otherwise specified.
- Do not write comments that summarize the code. Comments should only be
  written in order to explain "why" the code is written in some way in
  the case there is a reason that is tricky / non-obvious.
- Prefer implementing functionality in existing files unless it is a new
  logical component. Avoid creating many small files.
- Avoid using functions that panic like `unwrap()`, instead use
  mechanisms like `?` to propagate errors.
- Be careful with operations like indexing which may panic if the
  indexes are out of bounds.
- Never silently discard errors with `let _ =` on fallible operations.
- Avoid creative additions unless explicitly requested
- Use full words for variable names (no abbreviations like "q" for
  "queue")
- Avoid superfluous type annotations.
- Run tests and clippy after each change.
- Prefer inline snapshots with raw strings when writing test.
- Use `cargo run --` instead of the built binary when running the app
  manually.
