use anyhow::{bail, Context as _, Result};
use serde::Serialize;

use crate::report::{Comment, Severity};

#[derive(Serialize, Clone)]
struct ReviewComment {
    path: String,
    line: usize,
    body: String,
}

impl ReviewComment {
    fn from_comment(comment: &Comment, cwd: &str) -> Result<Self> {
        if comment.line == 0 {
            bail!("comment has line 0, which is invalid for GitHub review API: {comment:?}");
        }
        let prefix = format!("{cwd}/");
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
            body: format_body(comment.severity, &comment.message),
        })
    }
}

#[derive(Serialize)]
struct ReviewRequest {
    commit_id: String,
    body: String,
    event: String,
    comments: Vec<ReviewComment>,
}

fn format_body(severity: Severity, message: &str) -> String {
    let kind = match severity {
        Severity::Error => "CAUTION",
        Severity::Warning => "WARNING",
    };
    format!("**{kind}**: {}", html_escape::encode_text(message))
}

fn to_review_comments(
    comments: &[Comment],
    cwd: &str,
) -> Result<Vec<ReviewComment>> {
    comments
        .iter()
        .map(|comment| ReviewComment::from_comment(comment, cwd))
        .collect()
}

fn pr_from_ref(reference: &str) -> Option<u64> {
    reference
        .strip_prefix("refs/pull/")?
        .split('/')
        .next()?
        .parse()
        .ok()
}

const MAX_COMMENTS_PER_REVIEW: usize = 30;

fn send_batch(
    review_comments: &[ReviewComment],
    token: &str,
    repo: &str,
    pr: u64,
    sha: &str,
) -> Result<()> {
    let url =
        format!("https://api.github.com/repos/{repo}/pulls/{pr}/reviews");
    let body = ReviewRequest {
        commit_id: sha.to_string(),
        body: String::new(),
        event: "COMMENT".to_string(),
        comments: review_comments.to_vec(),
    };
    ureq::post(&url)
        .set("Authorization", &format!("Bearer {token}"))
        .set("Accept", "application/vnd.github+json")
        .send_json(&body)
        .context("posting review")?;
    Ok(())
}

pub fn run(comments: &[Comment]) -> Result<()> {
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

    let review_comments = to_review_comments(comments, &cwd)?;

    if review_comments.is_empty() {
        println!("no comments");
        return Ok(());
    }

    for batch in review_comments.chunks(MAX_COMMENTS_PER_REVIEW) {
        send_batch(batch, &token, &repo, pr, &sha)?;
    }
    Ok(())
}
