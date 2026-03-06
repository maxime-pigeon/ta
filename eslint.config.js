import jsdoc from "eslint-plugin-jsdoc";
import sonarjs from "eslint-plugin-sonarjs";
import unicorn from "eslint-plugin-unicorn";

export default [
    jsdoc.configs["flat/recommended"],
    {
        plugins: { jsdoc, sonarjs, unicorn },
        rules: {
            "no-unused-vars": "error",
            "eqeqeq": "error",
            "no-unreachable": "error",
            "no-redeclare": "error",
            "no-var": "error",
            "prefer-const": "error",
            "no-console": "error",
            "no-shadow": "error",
            "max-lines-per-function": ["warn", 20],
            "no-magic-numbers": ["warn", { "ignore": [0, 1, -1] }],
            "complexity": ["warn", 5],
            "max-depth": ["warn", 3],
            "no-nested-ternary": "error",
            "no-constant-condition": "error",
            "no-invalid-this": "error",
            "no-duplicate-case": "error",
            "default-case": "error",
            "no-fallthrough": "error",
            "consistent-return": "error",
            "no-else-return": "error",
            "no-implicit-globals": "error",
            "camelcase": "error",
            "func-style": ["error", "declaration"],
            "jsdoc/require-jsdoc": ["error", {
                "require": {
                    "FunctionDeclaration": true,
                    "ArrowFunctionExpression": true,
                },
            }],
            "jsdoc/require-param": "error",
            "jsdoc/require-param-type": "error",
            "jsdoc/require-returns": "error",
            "jsdoc/require-returns-type": "error",
            "prefer-template": "error",
            "no-restricted-properties": ["error", {
                "property": "innerHTML",
                "message": "Use textContent, createElement or append instead.",
            }],
            "sonarjs/no-duplicate-string": "error",
            "unicorn/prefer-query-selector": "error",
            "unicorn/prefer-add-event-listener": "error",
            "no-restricted-syntax": ["error",
                {
                    "selector": "Program > VariableDeclaration",
                    "message": "Do not declare variables at the top level. Move this inside a function or class.",
                },
                {
                    "selector": "FunctionDeclaration FunctionDeclaration",
                    "message": "Do not declare functions inside other functions. Move this function to the top level.",
                },
            ],
        },
    },
];
