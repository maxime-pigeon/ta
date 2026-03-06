// linters routes files to the appropriate linter based on their extension
// and normalizes results into a common Lint type.

import { ESLint } from "eslint";
import { HtmlValidate } from "html-validate";
import stylelint from "stylelint";

/** Lint is a single finding produced by a linter. */
export interface Lint {
    filePath: string;
    line: number;
    column: number;
    rule: string;
    message: string;
    severity: "error" | "warning";
}

const JS_EXT = /\.(js|ts|jsx|tsx|mjs|cjs)$/i;
const CSS_EXT = /\.(css|scss|less)$/i;
const HTML_EXT = /\.html?$/i;

/**
 * lint routes each file to the appropriate linter based on its extension
 * and returns all findings.
 */
export async function lint(files: string[]): Promise<Lint[]> {
    const js = files.filter((f) => JS_EXT.test(f));
    const css = files.filter((f) => CSS_EXT.test(f));
    const html = files.filter((f) => HTML_EXT.test(f));

    const results = await Promise.all([
        js.length > 0 ? lintWithEslint(js) : Promise.resolve([]),
        css.length > 0 ? lintWithStylelint(css) : Promise.resolve([]),
        html.length > 0 ? lintWithHtmlValidate(html) : Promise.resolve([]),
    ]);
    return results.flat();
}

/** lintWithEslint runs ESLint on the given files and returns findings. */
async function lintWithEslint(files: string[]): Promise<Lint[]> {
    const eslint = new ESLint();
    const results = await eslint.lintFiles(files);
    return results.flatMap((result) =>
        result.messages.map((msg) => ({
            filePath: result.filePath,
            line: msg.line,
            column: msg.column,
            rule: msg.ruleId ?? "unknown",
            message: msg.message,
            severity: msg.severity === 2 ? "error" : "warning",
        }))
    );
}

/** lintWithStylelint runs stylelint on the given files and returns findings. */
async function lintWithStylelint(files: string[]): Promise<Lint[]> {
    const result = await stylelint.lint({ files });
    return result.results.flatMap((res) =>
        res.warnings.map((w) => ({
            filePath: res.source ?? "",
            line: w.line,
            column: w.column,
            rule: w.rule,
            message: w.text,
            severity: w.severity === "error" ? "error" : "warning",
        }))
    );
}

/**
 * lintWithHtmlValidate runs html-validate on the given files and returns
 * findings.
 */
async function lintWithHtmlValidate(files: string[]): Promise<Lint[]> {
    const validator = new HtmlValidate();
    const lints: Lint[] = [];
    for (const file of files) {
        const report = await validator.validateFile(file);
        for (const result of report.results) {
            for (const msg of result.messages) {
                lints.push({
                    filePath: result.filePath,
                    line: msg.line,
                    column: msg.column,
                    rule: msg.ruleId,
                    message: msg.message,
                    severity: msg.severity === 2 ? "error" : "warning",
                });
            }
        }
    }
    return lints;
}

/** srcFiles returns all lintable files under the given directory. */
export async function srcFiles(dir: string): Promise<string[]> {
    const glob = new Bun.Glob(
        "**/*.{js,ts,jsx,tsx,mjs,cjs,css,scss,less,html,htm}",
    );
    return Array.fromAsync(glob.scan({ cwd: dir, absolute: true }));
}
