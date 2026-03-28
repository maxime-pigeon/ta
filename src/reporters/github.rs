use std::collections::HashSet;

use anyhow::{Context as _, Result};
use serde::{Deserialize, Serialize};

use crate::report::{Comment, Severity};

struct PostOptions {
    token: String,
    repo: String,
    pr: u64,
    sha: String,
}

#[derive(Serialize)]
struct ReviewComment {
    path: String,
    line: usize,
    body: String,
}

#[derive(Serialize)]
struct ReviewRequest<'a> {
    commit_id: &'a str,
    body: &'a str,
    event: &'a str,
    comments: &'a [ReviewComment],
}

#[derive(Deserialize)]
struct PullFile {
    filename: String,
}

fn format_body(severity: Severity, message: &str) -> String {
    let kind = match severity {
        Severity::Error => "CAUTION",
        Severity::Warning => "WARNING",
    };
    format!("> [!{}]\n> {}", kind, html_escape::encode_text(message))
}

fn to_review_comments(
    comments: &[Comment],
    cwd: &str,
    changed: &HashSet<String>,
) -> Vec<ReviewComment> {
    let prefix = format!("{cwd}/");
    comments
        .iter()
        .map(|c| {
            let path = c
                .filepath
                .strip_prefix(&prefix)
                .unwrap_or(&c.filepath)
                .to_string();
            ReviewComment {
                path,
                line: c.line,
                body: format_body(c.severity, &c.message),
            }
        })
        .filter(|rc| changed.contains(&rc.path))
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

fn get_changed_files(token: &str, repo: &str, pr: u64) -> Result<Vec<String>> {
    let url = format!("https://api.github.com/repos/{repo}/pulls/{pr}/files");
    let files: Vec<PullFile> = ureq::get(&url)
        .set("Authorization", &format!("Bearer {token}"))
        .set("Accept", "application/vnd.github+json")
        .call()
        .context("fetching changed files")?
        .into_json()
        .context("parsing changed files response")?;
    Ok(files.into_iter().map(|f| f.filename).collect())
}

fn post_review(
    review_comments: &[ReviewComment],
    opts: &PostOptions,
) -> Result<()> {
    let max_comments_per_request = 30;
    for batch in review_comments.chunks(max_comments_per_request) {
        send_batch(batch, opts)?;
    }
    Ok(())
}

fn send_batch(
    review_comments: &[ReviewComment],
    opts: &PostOptions,
) -> Result<()> {
    let url = format!(
        "https://api.github.com/repos/{}/pulls/{}/reviews",
        opts.repo, opts.pr
    );
    let body = ReviewRequest {
        commit_id: &opts.sha,
        body: "",
        event: "COMMENT",
        comments: review_comments,
    };
    ureq::post(&url)
        .set("Authorization", &format!("Bearer {}", opts.token))
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

    let changed: HashSet<String> =
        get_changed_files(&token, &repo, pr)?.into_iter().collect();
    let review_comments = to_review_comments(comments, &cwd, &changed);

    if review_comments.is_empty() {
        println!("no comments");
        return Ok(());
    }

    post_review(
        &review_comments,
        &PostOptions {
            token,
            repo,
            pr,
            sha,
        },
    )?;
    println!("posted review with {} comment(s)", review_comments.len());
    Ok(())
}
