use anyhow::{bail, Context as _, Result};
use itertools::process_results;
use serde::Deserialize;

use super::{Comment, Severity};

#[derive(Deserialize)]
struct FileEntry {
    source: String,
    warnings: Vec<Lint>,
}

#[derive(Deserialize)]
struct Lint {
    rule: String,
    severity: String,
    text: String,
    line: usize,
    column: usize,
}

/// Parses Stylelint JSON output into a list of lints.
pub fn parse(data: &str) -> Result<Vec<Comment>> {
    let file_entries: Vec<FileEntry> =
        serde_json::from_str(data).context("parsing Stylelint JSON")?;
    process_results(
        file_entries.into_iter().flat_map(|file_entry| {
            let source = file_entry.source.clone();
            file_entry
                .warnings
                .into_iter()
                .map(move |lint| lint.into_comment(source.clone()))
        }),
        |iter| iter.collect(),
    )
}

impl Lint {
    fn into_comment(self, filepath: String) -> Result<Comment> {
        let severity = match self.severity.as_str() {
            "error" => Severity::Error,
            "warning" => Severity::Warning,
            other => bail!("unknown Stylelint severity: {other}"),
        };
        Ok(Comment {
            filepath,
            line: self.line,
            col: self.column,
            severity,
            rule: self.rule,
            message: self.text,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_lint(severity: &str, text: &str) -> Lint {
        Lint {
            rule: "color-no-invalid-hex".to_string(),
            severity: severity.to_string(),
            text: text.to_string(),
            line: 2,
            column: 10,
        }
    }

    #[test]
    fn into_comment_error_severity() {
        let comment = make_lint(
            "error",
            "Unexpected invalid hex color (color-no-invalid-hex)",
        )
        .into_comment("src/style.css".to_string())
        .unwrap();
        insta::assert_debug_snapshot!(comment, @r#"
            Comment {
                filepath: "src/style.css",
                line: 2,
                col: 10,
                severity: Error,
                rule: "color-no-invalid-hex",
                message: "Unexpected invalid hex color (color-no-invalid-hex)",
            }
        "#);
    }

    #[test]
    fn into_comment_warning_severity() {
        let comment = make_lint(
            "warning",
            "Unexpected invalid hex color (color-no-invalid-hex)",
        )
        .into_comment("src/style.css".to_string())
        .unwrap();
        insta::assert_debug_snapshot!(comment, @r#"
            Comment {
                filepath: "src/style.css",
                line: 2,
                col: 10,
                severity: Warning,
                rule: "color-no-invalid-hex",
                message: "Unexpected invalid hex color (color-no-invalid-hex)",
            }
        "#);
    }

    #[test]
    fn into_comment_unknown_severity() {
        let error = make_lint("info", "some message")
            .into_comment("src/style.css".to_string())
            .unwrap_err();
        insta::assert_snapshot!(error.to_string(), @r#"
            unknown Stylelint severity: info
        "#);
    }
}
