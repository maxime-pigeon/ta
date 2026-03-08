// ta reads linter output by running linters programmatically on all files
// in the src folder and either prints findings to stdout or posts them as
// inline GitHub pull-request review comments.

import { getChangedFiles, postReview } from "./github.ts";
import { lint, toComments, type Lint } from "./linters.ts";

const isTTY = process.stdout.isTTY ?? false;
const bold = (s: string) => isTTY ? `\x1b[1m${s}\x1b[0m` : s;
const red = (s: string) => isTTY ? `\x1b[31m${s}\x1b[0m` : s;
const yellow = (s: string) => isTTY ? `\x1b[33m${s}\x1b[0m` : s;
const cyan = (s: string) => isTTY ? `\x1b[36m${s}\x1b[0m` : s;

/**
 * readLine reads a single line from a file. Returns "" on any error so
 * callers can degrade gracefully when the file is unreadable.
 */
async function readLine(
    filePath: string,
    lineNum: number,
): Promise<string> {
    try {
        const text = await Bun.file(filePath).text();
        return text.split("\n")[lineNum - 1] ?? "";
    } catch {
        return "";
    }
}

/**
 * formatLint formats a single Lint finding with its source line into a
 * multi-line diagnostic block similar to Rust's rustc output.
 */
function formatLint(l: Lint, sourceLine: string): string {
    const label = l.severity === "error" ? "Error" : "Warning";
    const color = l.severity === "error" ? red : yellow;
    const gutterW = String(l.line).length;
    const pad = " ".repeat(gutterW);
    const caret = " ".repeat(Math.max(0, l.column - 1)) + "^";
    const arrow = bold(cyan("-->"));
    const pipe = bold(cyan("|"));
    const heading = bold(color(`${label}: ${l.message}`));
    const lines = [
        heading,
        `  ${arrow} ${l.filePath}:${l.line}:${l.column}`,
        `   ${pad}${pipe}`,
    ];
    if (sourceLine) {
        lines.push(`${l.line.toString().padStart(gutterW + 2)} ${pipe} ${sourceLine}`);
        lines.push(`   ${pad}${pipe} ${caret}`);
    }
    return lines.join("\n");
}

/** srcFiles returns all lintable files under the given directory. */
async function srcFiles(dir: string): Promise<string[]> {
    const glob = new Bun.Glob(
        "**/*.{js,ts,jsx,tsx,mjs,cjs,css,scss,less,html,htm}",
    );
    return Array.fromAsync(glob.scan({ cwd: dir, absolute: true }));
}

/** runStdout lints src files and prints findings to stdout. */
async function runStdout(): Promise<void> {
    const files = await srcFiles(`${process.cwd()}/src`);
    const lints = await lint(files);
    const parts = await Promise.all(
        lints.map(async (l) =>
            formatLint(l, await readLine(l.filePath, l.line))
        ),
    );
    if (parts.length > 0) {
        console.log(parts.join("\n\n"));
    }
}

/**
 * prFromRef parses the pull-request number from a GITHUB_REF value such as
 * "refs/pull/42/merge", returning 0 if the ref does not match.
 */
function prFromRef(ref: string): number {
    const m = ref.match(/^refs\/pull\/(\d+)\//);
    return m ? parseInt(m[1], 10) : 0;
}

/** runCI lints src files and posts inline comments on the pull request. */
async function runCI(): Promise<void> {
    const token = process.env.GITHUB_TOKEN ?? "";
    const repo = process.env.GITHUB_REPOSITORY ?? "";
    const sha = process.env.GITHUB_SHA ?? "";
    const pr = parseInt(process.env.PR_NUMBER ?? "0", 10)
        || prFromRef(process.env.GITHUB_REF ?? "");

    const missing: string[] = [];
    if (!token) missing.push("GITHUB_TOKEN");
    if (!repo) missing.push("GITHUB_REPOSITORY");
    if (!sha) missing.push("GITHUB_SHA");
    if (!pr) {
        missing.push(
            "PR_NUMBER or GITHUB_REF=refs/pull/N/merge "
                + `(got ${JSON.stringify(process.env.GITHUB_REF ?? "")})`,
        );
    }
    if (missing.length > 0) {
        console.error(`CI mode: missing ${missing.join(", ")}`);
        process.exit(1);
    }

    const cwd = process.cwd();
    const changed = new Set(await getChangedFiles(token, repo, pr));
    const files = await srcFiles(`${cwd}/src`);
    const comments = toComments(await lint(files), "markdown")
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

const USAGE = `\
Usage: ta [--ci] [--help]

Lint all files under ./src and report findings.

Options:
  (none)   Print findings to stdout (default).
  --ci     Post inline comments on a GitHub pull-request review.
           Requires environment variables:
             GITHUB_TOKEN       Personal access token with repo scope.
             GITHUB_REPOSITORY  Owner/repo (e.g. "acme/myapp").
             GITHUB_SHA         Commit SHA being reviewed.
             GITHUB_REF         Set automatically on pull_request events
                                (refs/pull/N/merge). Used to derive the
                                PR number when PR_NUMBER is not set.
             PR_NUMBER          Pull-request number (overrides GITHUB_REF).
  --help   Print this message and exit.
`;

async function main(): Promise<void> {
    const args = process.argv.slice(2);

    if (args.includes("--help") || args.includes("-h")) {
        process.stdout.write(USAGE);
        return;
    }

    const mode = args.includes("--ci") ? "ci" : "stdout";

    switch (mode) {
        case "stdout":
            await runStdout();
            break;
        case "ci":
            await runCI();
            break;
        default:
            console.error(`unknown mode: ${mode}`);
            process.exit(1);
    }
}

main().catch((err: unknown) => {
    console.error(err instanceof Error ? err.message : err);
    process.exit(1);
});
