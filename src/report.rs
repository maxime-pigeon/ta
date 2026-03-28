use std::fs;

use anyhow::{Context as _, Error, Result};

#[derive(Clone, Copy, Debug)]
pub enum Severity {
    Error,
    Warning,
}

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

struct Linter {
    name: String,
    filepath: String,
    parse: fn(&str) -> Result<Vec<Comment>>,
}

pub struct Report {
    linters: Vec<Linter>,
    pub errors: Vec<Error>,
}

impl Report {
    pub fn new() -> Self {
        Report {
            linters: Vec::new(),
            errors: Vec::new(),
        }
    }

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

    pub fn build(&mut self) -> Result<Vec<Comment>> {
        let mut comments = Vec::new();
        let mut parse_errors = Vec::new();

        for linter in &self.linters {
            let json = fs::read_to_string(&linter.filepath)
                .with_context(|| format!("reading {}", linter.filepath))?;
            match (linter.parse)(&json) {
                Ok(lints) => comments.extend(lints),
                Err(err) => {
                    parse_errors.push(err.context(linter.name.clone()));
                }
            }
        }

        self.errors = parse_errors;

        comments.sort_by(|a, b| {
            a.filepath
                .cmp(&b.filepath)
                .then(a.line.cmp(&b.line))
                .then(a.col.cmp(&b.col))
        });

        Ok(comments)
    }
}
