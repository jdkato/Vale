# Contributing to Vale

Interested in contributing to Vale? Great&mdash;we welcome contributions of any kind including documentation improvements, bug reports, feature requests, and pull requests.

## Table of Contents

- [Contributing to Vale](#contributing-to-vale)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
  - [Testing](#testing)
  - [Setting up a Development Environment](#setting-up-a-development-environment)
  - [Code Contribution Guidelines](#code-contribution-guidelines)
  - [Git Commit Message Guidelines](#git-commit-message-guidelines)
  - [Terminology](#terminology)

## Introduction

Vale is a natural language linter that supports plain text, markup (Markdown, reStructuredText, AsciiDoc, and HTML), and source code comments. However, unlike many similar projects, Vale's primary focus isn't on providing a collection of rules everyone must follow&mdash;instead, Vale aims to be flexible enough to support many different styles (see [Styles](https://errata-ai.github.io/vale/styles/) for more information).

More specifically, Vale is written in Go and split into packages that are tasked with implementing specific functionality:

- `check` handles the loading and validating of external rules (YAML files).
- `core`: includes the main structures used throughout the application (e.g., `File` and `Alert`) and manages configuration files.
- `lint` handles the actual linting, which includes knowing when to apply rules and how to handle specific file formats.
- `rule` implements Vale's built-in style.
- `ui` manages displaying information to users.

If you're looking to improve Vale's documentation, check out the [`docs/`](https://github.com/errata-ai/vale/tree/master/docs) directory.

## Testing

Vale is tested using both integration and unit tests.

Integration tests are the most plentiful at the moment. They're implemented using the behavior-driven development framework [Cucumber](https://cucumber.io/). You'll find the relevant files for these tests in the `fixtures` and `features` directories. Unit tests are found in the `*_test.go` files inside the actual Go packages.

We also track Vale's performance on a per-commit basis through benchmarks. On every commit, you'll see comparison against the last tagged release (over 5 runs) on CI builds:

```text
LintRST-2   1.63s ± 2%   1.65s ± 2%  +0.95%  (p=0.031 n=10+10)
LintMD-2    1.54s ± 1%   1.54s ± 1%    ~     (p=0.604 n=10+10)
```


To run the tests, you'll want to invoke either `make bench` or `make ci` (see [Setting up a Development Environment]() for more information).

## Setting up a Development Environment

Prerequisites:

* [Ruby](https://www.ruby-lang.org/en/downloads/) (v2.3+)
* [Go](https://golang.org/) (v1.7+) installed.
* [Asciidoctor](http://asciidoctor.org/) available on your `$PATH`.
* [rst2html](http://docutils.sourceforge.net/docs/user/tools.html#rst2html-py) available on your `$PATH`. The latter is installed with both [Sphinx](http://www.sphinx-doc.org/en/stable/) and [docutils](https://pypi.python.org/pypi/docutils).
* [xsltproc](http://xmlsoft.org/xslt/xsltproc.html) available on your `$PATH`.
* [dita](https://www.dita-ot.org/download) available on your `$PATH` (v3.6+).

```bash
$ gem install bundler:2.2.31 # if necessary
$ make setup
$ make test
```

## Code Contribution Guidelines

To make the contribution process as seamless as possible, we ask for the following:

* Fork the project and make your changes.
* When you’re ready to create a pull request, be sure to:
    * Run `make lint` to check your Go code style.
    * Squash your commits into a single commit. `git rebase -i`. It’s okay to force update your pull request with `git push -f`.
    * Follow the **Git Commit Message Guidelines** below.

## Git Commit Message Guidelines

Vale follows a modified version of the [AngularJS Commit Guidelines](https://github.com/angular/angular.js/blob/master/CONTRIBUTING.md#-git-commit-guidelines). A commit message should take the following form:

```
<type>: <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

with `<body>` and `<footer>` being optional. `<type>` should be one of the following:

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes (e.g., this document, the README, or source comments)
- `style`: Changes that do not affect the meaning of the code (e.g., code formatting)
- refactor: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance (in this case, please include relevant benchmark(s))
- `test`: Adding missing or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

An example would be something like:

```text
refactor: make "warning" the default lint level

Also demotes `Annotations` and `PassiveVoice` to "suggestions."

Related to #30.
```

## Terminology

| Term  | Definition                                                                                                                                                                        |
|:-----:|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| check | A "check" is one of Vale's extension points (e.g., `existence` and `substitution`) that performs a single task such as looking for the existence of a word.                       |
| rule  | A "rule" is an actual implementation of a check. For example, [`Hedging`](https://github.com/errata-ai/vale/blob/master/rule/vale/Hedging.yml) is one of Vale's built-in rules.         |
| style | A "style" is a collection of rules. For example, [`Joblint`](https://github.com/errata-ai/vale/tree/master/rule/Joblint) is a style that consists of rules such as `LegacyTech`. |
