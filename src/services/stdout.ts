// stdout formats lint findings and prints them to the terminal.

import chalk from "chalk";
import { type Lint, lint, srcFiles } from "../linters.ts";

const ARROW = chalk.bold.cyan("-->");
const PIPE = chalk.bold.cyan("|");

/**
 * readFiles reads each file path once and returns a map of path to lines.
 */
async function readFiles(paths: string[]): Promise<Map<string, string[]>> {
    const unique = [...new Set(paths)];
    const entries = await Promise.all(
        unique.map(async (p): Promise<[string, string[]]> => {
            try {
                const text = await Bun.file(p).text();
                return [p, text.split("\n")];
            } catch (err) {
                throw new Error(`readFiles: cannot read ${p}`, {
                    cause: err,
                });
            }
        }),
    );
    return new Map(entries);
}

/**
 * formatLint formats a single Lint finding with its source line into a
 * multi-line diagnostic block.
 */
export function formatLint(l: Lint, sourceLine: string): string {
    const label = l.severity === "error" ? "Error" : "Warning";
    const color = l.severity === "error" ? chalk.red : chalk.yellow;
    const gutterW = String(l.line).length;
    const pad = " ".repeat(gutterW);
    const caret = " ".repeat(Math.max(0, l.column - 1)) + "^";
    const heading = chalk.bold(color(`${label}: ${l.message}`));
    const lineNo = l.line.toString().padStart(gutterW + 2);
    return [
        heading,
        `  ${ARROW} ${l.filePath}:${l.line}:${l.column}`,
        `   ${pad}${PIPE}`,
        `${lineNo} ${PIPE} ${sourceLine}`,
        `   ${pad}${PIPE} ${caret}`,
    ].join("\n");
}

/** runStdout lints src files and prints findings to stdout. */
export async function runStdout(): Promise<void> {
    const files = await srcFiles(`${process.cwd()}/src`);
    const lints = await lint(files);
    const fileLines = await readFiles(lints.map((l) => l.filePath));
    const formattedLints = lints.map((l) => {
        const lines = fileLines.get(l.filePath) ?? [];
        const sourceLine = lines[l.line - 1];
        if (sourceLine === undefined) {
            throw new Error(
                `line ${l.line} is out of bounds in ${l.filePath}`,
            );
        }
        return formatLint(l, sourceLine);
    });
    console.log(formattedLints.join("\n\n"));
}
