//! Parsers for linter JSON output formats.

pub mod eslint;
pub mod stylelint;

pub use crate::report::{Comment, Severity};
