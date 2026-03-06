// no-var: use let or const instead of var
var unusedVar = 42;

// no-unused-vars: variable declared but never used
const unusedConst = "hello";

// prefer-const: let used but never reassigned
let neverReassigned = 10;

// eqeqeq: use === instead of ==
if (neverReassigned == null) {}

// no-console: console statements not allowed
console.log("debug");

// no-unreachable: code after return
function unreachableCode() {
    return true;
    console.log("never reached");
}

// no-redeclare: variable declared twice
var redeclared = "first";
var redeclared = "second";

// no-shadow: inner variable shadows outer
let outerCount = 0;
function shadowExample() {
    let outerCount = 1;
    return outerCount;
}

// func-style: use function declarations, not expressions
const greetFn = function(name) {
    return "Hello " + name;
};

// jsdoc/require-jsdoc: missing JSDoc comment
function undocumented() {
    return true;
}

// camelcase: use camelCase for variable names
const my_var = "value";

// no-magic-numbers: avoid unnamed numeric constants (warn)
function magicNumbers() {
    return 99 * 3;
}

// max-lines-per-function: function exceeds 20 lines (warn)
function tooLong() {
    const l1 = 1;
    const l2 = 2;
    const l3 = 3;
    const l4 = 4;
    const l5 = 5;
    const l6 = 6;
    const l7 = 7;
    const l8 = 8;
    const l9 = 9;
    const l10 = 10;
    const l11 = 11;
    const l12 = 12;
    const l13 = 13;
    const l14 = 14;
    const l15 = 15;
    const l16 = 16;
    const l17 = 17;
    const l18 = 18;
    const l19 = 19;
    const l20 = 20;
    return l1 + l2 + l3 + l4 + l5 + l6 + l7 + l8 + l9 + l10 + l11 + l12 + l13 + l14 + l15 + l16 + l17 + l18 + l19 + l20;
}

// complexity: too many branches (warn)
function tooComplex(a, b, c, d, e, f) {
    if (a) { return 1; }
    if (b) { return 2; }
    if (c) { return 3; }
    if (d) { return 4; }
    if (e) { return 5; }
    if (f) { return 6; }
    return 0;
}

// max-depth: too many levels of nesting (warn)
function tooDeep(a, b, c, d) {
    if (a) {
        if (b) {
            if (c) {
                if (d) {
                    return true;
                }
            }
        }
    }
    return false;
}

// no-nested-ternary: ternary inside a ternary
const score = 75;
const grade = score >= 90 ? "A" : score >= 70 ? "B" : "C";

// no-constant-condition: condition is always true
if (true) {
    const alwaysRuns = 1;
}

// no-invalid-this: this used outside of a class or object method
function standaloneFunction() {
    return this.value;
}

// no-duplicate-case: same case value appears twice
switch (unusedVar) {
    case 1:
        break;
    case 1:
        break;
    default:
        break;
}

// default-case: switch without a default clause
switch (unusedVar) {
    case 1:
        break;
    case 2:
        break;
}

// no-fallthrough: case falls through to the next without break
switch (unusedVar) {
    case 1:
        outerCount = 1;
    case 2:
        outerCount = 2;
        break;
    default:
        break;
}

// consistent-return: sometimes returns a value, sometimes not
function sometimesReturns(x) {
    if (x > 0) {
        return x;
    }
}

// no-else-return: unnecessary else after a return
function withUnnecessaryElse(x) {
    if (x > 0) {
        return x;
    } else {
        return -x;
    }
}

// prefer-template: use template literals instead of string concatenation
const name = "world";
const greeting = "Hello " + name + "!";

// jsdoc/require-param: missing @param tag
/**
 * Adds two numbers.
 * @returns {number} The sum.
 */
function missingParam(a, b) {
    return a + b;
}

// jsdoc/require-returns: missing @returns tag
/**
 * Doubles a number.
 * @param {number} n - The number to double.
 */
function missingReturns(n) {
    return n * 2;
}

// jsdoc/require-param-type: missing type on @param
/**
 * Triples a number.
 * @param n - The number to triple.
 * @returns {number} The result.
 */
function missingParamType(n) {
    return n * 3;
}

// jsdoc/require-returns-type: missing type on @returns
/**
 * Negates a number.
 * @param {number} n - The number to negate.
 * @returns The negated value.
 */
function missingReturnsType(n) {
    return -n;
}

// sonarjs/no-duplicate-string: same string repeated 3+ times
const s1 = "some-duplicate-string";
const s2 = "some-duplicate-string";
const s3 = "some-duplicate-string";

// unicorn/prefer-query-selector: use querySelector instead of getElementById
const el1 = document.getElementById("app");
const el2 = document.getElementsByClassName("btn");

// no-restricted-properties: innerHTML is disallowed
const div = document.querySelector("div");
div.innerHTML = "<p>hello</p>";

// unicorn/prefer-add-event-listener: use addEventListener instead of onclick
const btn = document.querySelector("button");
btn.onclick = function() {};

// no-restricted-syntax (Program > VariableDeclaration): top-level variable declaration
const topLevel = { key: "value" };

// no-restricted-syntax (FunctionDeclaration FunctionDeclaration): nested function declaration
function outer() {
    function inner() {
        return true;
    }
    return inner();
}

// no-implicit-globals: n/a — does not apply to ES modules
