// github posts inline pull-request review comments via the GitHub REST API.

import { type Lint, lint, srcFiles } from "../linters.ts";

/** Comment is a student-facing piece of feedback derived from a Lint. */
export interface Comment {
    path: string;
    line: number;
    body: string;
}

/** PostOptions holds the credentials and target for a GitHub PR review. */
export interface PostOptions {
    token: string;
    repo: string;
    pr: number;
    sha: string;
}

interface ReviewRequest {
    commit_id: string;
    body: string;
    event: string;
    comments: Comment[];
}

/**
 * formatBody formats a severity and message into a GitHub callout.
 */
export function formatBody(
    severity: "error" | "warning",
    message: string,
): string {
    const type = severity === "error" ? "CAUTION" : "WARNING";
    return `> [!${type}]\n> ${Bun.escapeHTML(message)}`;
}

/** toComments converts Lints into student-facing Comments. */
export function toComments(lints: Lint[]): Comment[] {
    return lints.map((l) => ({
        path: l.filePath,
        line: l.line,
        body: formatBody(l.severity, l.message),
    }));
}

/**
 * prFromRef parses the pull-request number from a GITHUB_REF value such as
 * "refs/pull/42/merge", returning 0 if the ref does not match.
 */
export function prFromRef(ref: string): number {
    const m = ref.match(/^refs\/pull\/(\d+)\//);
    return m ? parseInt(m[1], 10) : 0;
}

/**
 * getChangedFiles returns the list of files changed in a pull request
 * via GET /repos/{owner}/{repo}/pulls/{pr}/files.
 */
export async function getChangedFiles(
    token: string,
    repo: string,
    pr: number,
): Promise<string[]> {
    const url = `https://api.github.com/repos/${repo}/pulls/${pr}/files`;
    const resp = await fetch(url, {
        headers: {
            Authorization: `Bearer ${token}`,
            Accept: "application/vnd.github+json",
        },
    });
    if (!resp.ok) {
        const body = await resp.text();
        throw new Error(`GitHub API error ${resp.status}: ${body}`);
    }
    const data: Array<{ filename: string }> = await resp.json();
    return data.map((f) => f.filename);
}

/**
 * postReview creates a GitHub pull-request review with one inline comment
 * per finding via POST /repos/{owner}/{repo}/pulls/{pr}/reviews.
 * Comments are sent in batches of 30 to avoid GitHub rate limits.
 */
export async function postReview(
    comments: Comment[],
    opts: PostOptions,
): Promise<void> {
    const maxPerRequest = 30;
    for (let i = 0; i < comments.length; i += maxPerRequest) {
        const batch = comments.slice(i, i + maxPerRequest);
        await sendBatch(batch, opts);
    }
}

/** sendBatch sends a single review request to the GitHub API. */
async function sendBatch(
    comments: Comment[],
    opts: PostOptions,
): Promise<void> {
    const url =
        `https://api.github.com/repos/${opts.repo}/pulls/${opts.pr}/reviews`;
    const resp = await fetch(url, {
        method: "POST",
        headers: {
            Authorization: `Bearer ${opts.token}`,
            Accept: "application/vnd.github+json",
            "Content-Type": "application/json",
        },
        body: JSON.stringify(buildReviewRequest(opts.sha, comments)),
    });
    if (!resp.ok) {
        const body = await resp.text();
        throw new Error(`GitHub API error ${resp.status}: ${body}`);
    }
}

function buildReviewRequest(
    sha: string,
    comments: Comment[],
): ReviewRequest {
    return {
        commit_id: sha,
        body: "",
        event: "COMMENT",
        comments,
    };
}

/** runGitHub lints src files and posts inline comments on the pull request. */
export async function runGitHub(): Promise<void> {
    const token = process.env.GITHUB_TOKEN;
    const repo = process.env.GITHUB_REPOSITORY;
    const sha = process.env.GITHUB_SHA;
    const ref = process.env.GITHUB_REF;

    if (!token || !repo || !sha || !ref) {
        console.log("missing GitHub environment variables");
        process.exit(1);
    }

    const pr = prFromRef(ref);

    const cwd = process.cwd();
    const changed = new Set(await getChangedFiles(token, repo, pr));
    const files = await srcFiles(`${cwd}/src`);
    const comments = toComments(await lint(files))
        .map((c) => ({
            ...c,
            path: c.path.startsWith(`${cwd}/`)
                ? c.path.slice(cwd.length + 1)
                : c.path,
        }))
        .filter((c) => changed.has(c.path));

    if (comments.length === 0) {
        console.log("no comments");
        return;
    }

    await postReview(comments, { token, repo, pr, sha });
    console.log(`posted review with ${comments.length} comment(s)`);
}
