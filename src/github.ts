import type { Comment } from "./linters.ts";

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

function buildReviewRequest(
    sha: string,
    comments: Comment[],
): ReviewRequest {
    return {
        commit_id: sha,
        body: "linter remarks",
        event: "COMMENT",
        comments,
    };
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
 * post creates a GitHub pull-request review with one inline comment per
 * finding via POST /repos/{owner}/{repo}/pulls/{pr}/reviews.
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
