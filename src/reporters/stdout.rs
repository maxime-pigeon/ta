use std::io::{self, Write};

use crate::report::{Comment, Severity};

pub fn print(comments: &[Comment]) {
    let mut stdout = io::stdout();
    for comment in comments {
        write!(stdout, "{}", comment_to_string(comment)).ok();
    }
}

fn comment_to_string(comment: &Comment) -> String {
    let severity = match comment.severity {
        Severity::Error => "error",
        Severity::Warning => "warning",
    };
    format!(
        "{}:{}:{}: {}: {} ({})\n",
        comment.filepath,
        comment.line,
        comment.col,
        severity,
        comment.message,
        comment.rule,
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn error_comment() {
        let comment = Comment::new(
            "src/main.rs",
            10,
            5,
            Severity::Error,
            "no-unused-vars",
            "x is defined but never used",
        );
        insta::assert_snapshot!(comment_to_string(&comment), @r#"
            src/main.rs:10:5: error: x is defined but never used (no-unused-vars)
        "#);
    }

    #[test]
    fn warning_comment() {
        let comment = Comment::new(
            "src/lib.rs",
            3,
            1,
            Severity::Warning,
            "prefer-const",
            "use const instead of let",
        );
        insta::assert_snapshot!(comment_to_string(&comment), @r#"
            src/lib.rs:3:1: warning: use const instead of let (prefer-const)
        "#);
    }
}
