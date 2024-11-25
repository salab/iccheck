# ICCheck - Inconsistent Change Checker

ICCheck takes any 2 revisions (including commit and worktree) from a Git repository
and reports possible inconsistent changes in pre-commit files, commited files, and/or Pull Requests.

ICCheck lists (pre-commit) changes made on a git repository
and checks missing changes on cloned codes (i.e. copy-pasted codes).

The plugin and the method for detecting cloned code is still in development and being improved.
While ICCheck detects many cloned codes correctly, note that it may also detect many false-positives.

## Installation

### Command Line Interface (CLI, Binary File)

- Download from the [latest releases](https://github.com/salab/iccheck/releases) page.
- Or, build it from source: `go install github.com/salab/iccheck@latest`

### Editor Extensions (VSCode, IntelliJ IDEA)

ICCheck utilizes [LSP (Language Server Protocol)](https://microsoft.github.io/language-server-protocol/) to support many editors with ease.

Currently, the following extensions are available:

- VSCode: [iccheck - Visual Studio Marketplace](https://marketplace.visualstudio.com/items?itemName=motoki317.iccheck)
- IntelliJ IDEA Ultimate: [ICCheck - Inconsistency Check - IntelliJ IDEs Plugin | Marketplace](https://plugins.jetbrains.com/plugin/24779-iccheck--inconsistency-check)

## Usage

### CLI

#### Input Format

Running `iccheck --help` displays help message.

```text
Finds inconsistent changes in your git changes.

Specify special values in base or target git ref arguments to compare against some special filesystems.
  "WORKTREE" : Compare against the current worktree.

Usage:
  iccheck [flags]
  iccheck [command]

Available Commands:
  help        Help about any command
  lsp         Starts ICCheck Language Server

Flags:
      --disable-default-ignore   Disable default ignore configs
      --fail-code int            Exit code if it detects any inconsistent changes (default: 0)
      --format string            Format type (console, json, github) (default "console")
  -f, --from string              Base git ref to compare against. Usually earlier in time. (default "main")
  -h, --help                     help for iccheck
      --ignore stringArray       Regexp of file paths (and its contents) to ignore.
                                 If specifying both file paths and contents ignore regexp, split them by ':'.
                                 Example (ignore dist directory): --ignore '^dist/'
                                 Example (ignore import statements in js files): --ignore '\.m?[jt]s$:^import'
      --log-level string         Log level (debug, info, warn, error) (default "info")
  -r, --repo string              Source git directory (default ".")
      --timeout-seconds int      Timeout for detecting clones in seconds (default: 15) (default 15)
  -t, --to string                Target git ref to compare from. Usually later in time. (default "HEAD")
  -v, --version                  version for iccheck

Use "iccheck [command] --help" for more information about a command.
```

Example:
Run ICCheck on this git repository for the last commit, to detect any inconsistent changes.

`iccheck --from HEAD~ --to HEAD --repo .`

#### Output Format

ICCheck outputs detected inconsistent changes to stdout, and other logging outputs to stderr.

Output format can be changed via the `--format` argument.
Make sure to check `--format json` out for ease integration with other systems such as review bots.

For example, one can utilize `jq` to process the JSON stdout into [the GitHub Actions annotation format](https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#example-creating-an-annotation-for-an-error).

```shell
iccheck --format json | jq -r '":::notice file=\(.filename),line=\(.start_l),endLine=\(.end_l),title=Possible missing change::Possible missing a consistent change here (L\(.start_l) - L\(.end_l), distance \(.distance))"'
```

#### In GitHub Actions

An example workflow file:

```yaml
name: Change Check

on:
  push:
    branches:
      - 'main'
  pull_request:

env:
  ICCHECK_FROM: "origin/main"
  ICCHECK_TO: "HEAD"

jobs:
  iccheck:
    name: Change Check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: '0'
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Set different base commit on main branch
        if: github.ref == 'refs/heads/main'
        run: echo "ICCHECK_FROM=HEAD~" >> "$GITHUB_ENV"
      - run: go install github.com/salab/iccheck@latest
      - run: iccheck --from "$ICCHECK_FROM" --to "$ICCHECK_TO" --format github
```

### Editor Extensions

Install the extension.
Then, edit any text files in a git-controlled directory.
ICCheck will automatically run when you open or edit files, and display line warnings
if you are likely missing changes to other similar lines.

![](./docs/editor-warning-example.png)

You can set cursor to warnings and run 'Find References' to display all clone
locations in the clone set.
(Shift+F12 in VSCode, Alt+F7 in IntelliJ)

![](./docs/find-references.png)

## Ignore Definitions

ICCheck reads from the following files to determine which files and/or lines to ignore,
when calculating info / warnings.

1. `~/.config/.iccheckignore.{yaml,yml}`
2. `${repoDir}/.iccheckignore.{yaml,yml}`

The ignore definitions is a yaml array and syntax is as follows:

```yaml
# The whole file should be a yaml array, each of which is a single ignore definition rule.

# "files" (if present) specifies which file paths to ignore.
# "patterns" (if present) specifies which file contents to ignore.
# Both keys use the golang regexp syntax: RE2
# For RE2 syntax, see: https://github.com/google/re2/wiki/syntax

# A rule to ignore all files under dist/ path.
- files:
    - "^dist/"

# A rule to ignore lines which match '^import' in any file.
- patterns:
    - '^import'

# A rule to ignore specific lines in .go files.
- files:
    - "\.go$"
  patterns:
    - '^package .+$'
    - '^import (.+ )?".+"$'
    - '^import \(\n(\s+(.+ )?".+"\n)*\)$'
```

There are some built-in default rules in ICCheck.
See [pkg/domain/ignore.go](./pkg/domain/ignore.go) for the default rules.
These default rules can be disabled by the `--disable-default-ignore` CLI option.
