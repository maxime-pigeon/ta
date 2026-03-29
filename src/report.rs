use std::fs;

use anyhow::{Context as _, Error, Result};

/// Severity level of a comment.
#[derive(Clone, Copy, Debug)]
pub enum Severity {
    Error,
    Warning,
}

/// A comment in a report.
#[derive(Debug)]
pub struct Comment {
    pub filepath: String,
    pub line: usize,
    pub col: usize,
    pub severity: Severity,
    pub rule: String,
    pub message: String,
}

#[cfg(test)]
impl Comment {
    pub(crate) fn new(
        filepath: &str,
        line: usize,
        col: usize,
        severity: Severity,
        rule: &str,
        message: &str,
    ) -> Self {
        Comment {
            filepath: filepath.to_string(),
            line,
            col,
            severity,
            rule: rule.to_string(),
            message: message.to_string(),
        }
    }
}

/// A registered linter with its output file path and parse function.
struct Linter {
    name: String,
    filepath: String,
    parse: fn(&str) -> Result<Vec<Comment>>,
}

/// Aggregates output from multiple linters into a unified list of comments.
pub struct Report {
    linters: Vec<Linter>,
    /// Parse errors from linters. Populated after calling [`Report::build`].
    pub errors: Vec<Error>,
}

impl Report {
    /// Creates a new report.
    pub fn new() -> Self {
        Report {
            linters: Vec::new(),
            errors: Vec::new(),
        }
    }

    /// Registers a linter whose output will be included when
    /// [`Report::build`] is called.
    pub fn add_linter(
        &mut self,
        name: &str,
        filepath: &str,
        parse: fn(&str) -> Result<Vec<Comment>>,
    ) {
        self.linters.push(Linter {
            name: name.to_string(),
            filepath: filepath.to_string(),
            parse,
        });
    }

    /// Reads and parses all registered linters, returning the combined
    /// comments.
    ///
    /// Returns an error if a linter's output file cannot be read. Linters
    /// whose output fails to parse are skipped; their errors are stored in
    /// [`Report::errors`].
    pub fn build(&mut self) -> Result<Vec<Comment>> {
        let mut comments = Vec::new();
        let mut parse_errors = Vec::new();
        for linter in &self.linters {
            let json = fs::read_to_string(&linter.filepath)
                .with_context(|| format!("reading {}", linter.filepath))?;
            match (linter.parse)(&json) {
                Ok(lints) => comments.extend(lints),
                // Parse errors are collected rather than returned early
                // so that a failure in one linter does not suppress
                // results from others.
                Err(err) => {
                    parse_errors.push(err.context(linter.name.clone()));
                }
            }
        }
        self.errors = parse_errors;
        Ok(comments)
    }
}
