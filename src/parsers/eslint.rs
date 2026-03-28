use anyhow::{bail, Context as _, Result};
use itertools::process_results;
use serde::Deserialize;

use super::{Comment, Severity};

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct FileEntry {
    file_path: String,
    messages: Vec<Lint>,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct Lint {
    rule_id: Option<String>,
    severity: u8,
    message: String,
    line: usize,
    column: usize,
}

pub fn parse(json: &str) -> Result<Vec<Comment>> {
    let files: Vec<FileEntry> =
        serde_json::from_str(json).context("parsing ESLint JSON")?;
    process_results(
        files.into_iter().flat_map(|file| {
            file.messages
                .into_iter()
                .map(move |lint| lint.into_comment(file.file_path.clone()))
        }),
        |iter| iter.collect(),
    )
}

impl Lint {
    fn into_comment(self, filepath: String) -> Result<Comment> {
        let severity = match self.severity {
            2 => Severity::Error,
            1 => Severity::Warning,
            other => bail!("unknown ESLint severity: {other}"),
        };
        Ok(Comment {
            filepath,
            line: self.line,
            col: self.column,
            severity,
            rule: self.rule_id.unwrap_or_default(),
            message: self.message,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_lint(severity: u8, rule_id: Option<&str>) -> Lint {
        Lint {
            rule_id: rule_id.map(String::from),
            severity,
            message: "Unexpected var.".to_string(),
            line: 3,
            column: 5,
        }
    }

    #[test]
    fn into_comment_error_severity() {
        let comment = make_lint(2, Some("no-var"))
            .into_comment("src/index.js".to_string())
            .unwrap();
        insta::assert_debug_snapshot!(comment, @r#"
            Comment {
                filepath: "src/index.js",
                line: 3,
                col: 5,
                severity: Error,
                rule: "no-var",
                message: "Unexpected var.",
            }
        "#);
    }

    #[test]
    fn into_comment_warning_severity() {
        let comment = make_lint(1, Some("no-console"))
            .into_comment("src/app.js".to_string())
            .unwrap();
        insta::assert_debug_snapshot!(comment, @r#"
            Comment {
                filepath: "src/app.js",
                line: 3,
                col: 5,
                severity: Warning,
                rule: "no-console",
                message: "Unexpected var.",
            }
        "#);
    }

    #[test]
    fn into_comment_missing_rule_id() {
        let comment = make_lint(2, None)
            .into_comment("src/index.js".to_string())
            .unwrap();
        insta::assert_debug_snapshot!(comment, @r#"
            Comment {
                filepath: "src/index.js",
                line: 3,
                col: 5,
                severity: Error,
                rule: "",
                message: "Unexpected var.",
            }
        "#);
    }

    #[test]
    fn into_comment_unknown_severity() {
        let error = make_lint(0, Some("no-var"))
            .into_comment("src/index.js".to_string())
            .unwrap_err();
        insta::assert_snapshot!(error.to_string(), @r#"
            unknown ESLint severity: 0
        "#);
    }
}
