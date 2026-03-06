// ta reads linter output by running linters programmatically on all files
// in the src folder and either prints findings to stdout or posts them as
// inline GitHub pull-request review comments.

import { Command } from "commander";
import { runGitHub } from "./services/github.ts";
import { runStdout } from "./services/stdout.ts";

const program = new Command();

program
    .name("ta")
    .description(
        "Lint all files under ./src and report findings.",
    )
    .option("--github", "Post inline comments on a GitHub PR review.")
    .action(async (opts: { github?: boolean }) => {
        if (opts.github) {
            await runGitHub();
        } else {
            await runStdout();
        }
    });

program.parseAsync(process.argv).catch((err: unknown) => {
    console.error(err instanceof Error ? err.message : err);
    process.exit(1);
});
