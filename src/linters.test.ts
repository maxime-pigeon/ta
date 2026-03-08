import { describe, it, expect } from "bun:test";
import { formatBody, toComments } from "./linters.ts";
import type { Lint } from "./linters.ts";

describe("formatBody", () => {
  describe("plain (default)", () => {
    it("prefixes error", () => {
      expect(formatBody("error", "something went wrong")).toBe(
        "Error: something went wrong",
      );
    });

    it("prefixes warning", () => {
      expect(formatBody("warning", "be careful")).toBe(
        "Warning: be careful",
      );
    });

    it("escapes html", () => {
      expect(formatBody("error", 'use <T> not "any"')).toBe(
        'Error: use &lt;T&gt; not &quot;any&quot;',
      );
    });
  });

  describe("markdown", () => {
    it("uses caution callout for error", () => {
      expect(formatBody("error", "something went wrong", "markdown")).toBe(
        "> [!CAUTION]\n> something went wrong",
      );
    });

    it("uses warning callout for warning", () => {
      expect(formatBody("warning", "be careful", "markdown")).toBe(
        "> [!WARNING]\n> be careful",
      );
    });

    it("escapes html", () => {
      expect(formatBody("error", 'use <T> not "any"', "markdown")).toBe(
        "> [!CAUTION]\n> use &lt;T&gt; not &quot;any&quot;",
      );
    });
  });
});

describe("toComments", () => {
  it("returns empty array for empty input", () => {
    expect(toComments([])).toEqual([]);
  });

  it("converts a single lint", () => {
    const lint: Lint = {
      filePath: "src/index.js",
      line: 5,
      column: 2,
      rule: "no-var",
      message: "Unexpected var.",
      severity: "error",
    };
    expect(toComments([lint])).toEqual([
      { path: "src/index.js", line: 5, body: "Error: Unexpected var." },
    ]);
  });

  it("preserves order", () => {
    const lints: Lint[] = [
      {
        filePath: "a.js",
        line: 1,
        column: 0,
        rule: "r1",
        message: "first",
        severity: "error",
      },
      {
        filePath: "b.js",
        line: 2,
        column: 0,
        rule: "r2",
        message: "second",
        severity: "warning",
      },
    ];
    const comments = toComments(lints);
    expect(comments).toHaveLength(2);
    expect(comments[0]).toEqual({
      path: "a.js",
      line: 1,
      body: "Error: first",
    });
    expect(comments[1]).toEqual({
      path: "b.js",
      line: 2,
      body: "Warning: second",
    });
  });
});
