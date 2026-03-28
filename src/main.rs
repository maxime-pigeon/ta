mod parsers;
mod report;
mod reporters;

use std::process;

use clap::{Parser, ValueEnum};
use report::Report;

#[derive(Parser)]
#[command(arg_required_else_help = true)]
struct Args {
    /// Path to ESLint JSON output
    #[arg(long, value_name = "PATH")]
    eslint: Option<String>,

    /// Path to Stylelint JSON output
    #[arg(long, value_name = "PATH")]
    stylelint: Option<String>,

    /// Path to html-validate JSON output
    #[arg(long = "html-validate", value_name = "PATH")]
    html_validate: Option<String>,

    /// Reporter to use
    #[arg(long, short, value_name = "REPORTER", default_value = "stdout")]
    reporter: Reporter,
}

#[derive(ValueEnum, Clone)]
enum Reporter {
    Stdout,
    Github,
}

fn main() {
    let args = Args::parse();

    let mut report = Report::new();
    if let Some(filepath) = &args.eslint {
        report.add_linter("ESLint", filepath, parsers::eslint::parse);
    }
    if let Some(filepath) = &args.stylelint {
        report.add_linter("Stylelint", filepath, parsers::stylelint::parse);
    }
    if let Some(filepath) = &args.html_validate {
        report.add_linter("html-validate", filepath, parsers::eslint::parse);
    }

    let comments = match report.build() {
        Ok(comments) => comments,
        Err(err) => {
            eprintln!("{err:#}");
            process::exit(1);
        }
    };

    for err in &report.errors {
        eprintln!("{err:#}");
    }

    match args.reporter {
        Reporter::Github => {
            if let Err(err) = reporters::github::run(&comments) {
                eprintln!("{err:#}");
                process::exit(1);
            }
        }
        Reporter::Stdout => reporters::stdout::print(&comments),
    }
}
