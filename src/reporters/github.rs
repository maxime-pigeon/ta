//! Reporter that posts comments as a GitHub pull request review.

use anyhow::{bail, Context as _, Result};
use serde::Serialize;

use crate::report::{Comment, Severity};

/// A single inline comment in a GitHub pull request review.
#[derive(Serialize)]
struct ReviewComment {
    path: String,
    line: usize,
    body: String,
}

impl ReviewComment {
    /// Converts a [`Comment`] into a [`ReviewComment`].
    fn from_comment(comment: &Comment, cwd: &str) -> Result<Self> {
        if comment.line == 0 {
            bail!("comment has line 0, which is invalid for GitHub review API: {comment:?}");
        }
        let prefix = format!("{cwd}/");
        // The GitHub review API requires paths relative to the repo
        // root, but linters report absolute paths.
        let path = comment
            .filepath
            .strip_prefix(&prefix)
            .with_context(|| {
                format!(
                    "filepath {:?} is not under current directory {cwd:?}",
                    comment.filepath
                )
            })?
            .to_string();
        Ok(Self {
            path,
            line: comment.line,
            body: format_comment_body(comment.severity, &comment.message),
        })
    }
}

/// The request body for the GitHub pull request review API.
#[derive(Serialize)]
struct ReviewRequest {
    commit_id: String,
    body: String,
    event: String,
    comments: Vec<ReviewComment>,
}

/// Formats a comment body with severity prefix and HTML-escaped message.
fn format_comment_body(severity: Severity, message: &str) -> String {
    let kind = match severity {
        Severity::Error => "error",
        Severity::Warning => "warning",
    };
    format!("TA {kind}: {}", html_escape::encode_text(message))
}

// Parses a PR number from a Git ref of the form `refs/pull/<number>/merge`.
fn pr_from_ref(reference: &str) -> Option<u64> {
    reference
        .strip_prefix("refs/pull/")?
        .split('/')
        .next()?
        .parse()
        .ok()
}

/// Posts lint comments as a GitHub pull request review.
pub fn post_review(comments: &[Comment]) -> Result<()> {
    let token =
        std::env::var("GITHUB_TOKEN").context("GITHUB_TOKEN not set")?;
    let repo = std::env::var("GITHUB_REPOSITORY")
        .context("GITHUB_REPOSITORY not set")?;
    let sha = std::env::var("GITHUB_SHA").context("GITHUB_SHA not set")?;
    let reference =
        std::env::var("GITHUB_REF").context("GITHUB_REF not set")?;

    let pr = pr_from_ref(&reference).with_context(|| {
        format!("could not parse PR number from GITHUB_REF: {reference}")
    })?;

    let cwd = std::env::current_dir()
        .context("getting current directory")?
        .to_string_lossy()
        .into_owned();

    let review_comments: Vec<ReviewComment> = comments
        .iter()
        .map(|comment| ReviewComment::from_comment(comment, &cwd))
        .collect::<Result<Vec<_>>>()?;

    if review_comments.is_empty() {
        return Ok(());
    }

    let url =
        format!("https://api.github.com/repos/{repo}/pulls/{pr}/reviews");
    let body = ReviewRequest {
        commit_id: sha.clone(),
        body: String::new(),
        event: "COMMENT".to_string(),
        comments: review_comments,
    };
    match ureq::post(&url)
        .set("Authorization", &format!("Bearer {token}"))
        .set("Accept", "application/vnd.github+json")
        .send_json(&body)
    {
        Ok(_) => {}
        Err(ureq::Error::Status(code, response)) => {
            let body = response.into_string().unwrap_or_else(|error| {
                format!("failed to read response: {error}")
            });
            bail!("GitHub API returned {code}: {body}");
        }
        Err(error) => return Err(error).context("posting review"),
    }

    Ok(())
}
