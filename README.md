# ICCheck - Inconsistent Change Checker

Reports possible inconsistent changes in pre-commit files, commited files, and/or Pull Requests.

ICCheck lists (pre-commit) changes made on a git repository
and checks missing changes on cloned codes (i.e. copy-pasted codes).

The plugin and the method for detecting cloned code is still in development and being improved.
While ICCheck detects many cloned codes correctly, note that it may also detect many false-positives.

## Installation

To be filled

## Usage

```text
Finds inconsistent changes in your git changes

Usage:
  iccheck [flags]

Flags:
      --fail-code int      Exit code if it detects any inconsistent changes (default: 0)
      --format string      Format type (console, json, github) (default "console")
  -f, --from string        Target git ref to compare against. Usually earlier in time. (default "main")
  -h, --help               help for iccheck
      --log-level string   Log level (debug, info, warn, error) (default "info")
  -r, --repo string        Source git directory (default ".")
  -t, --to string          Source git ref to compare from. Usually later in time. Set to 'WORKTREE' to specify worktree. (default "HEAD")
```

### Output Format

ICCheck outputs detected inconsistent changes to stdout, and other logging outputs to stderr.

Output format can be changed via the `--format` argument.
Make sure to check `--format json` out for ease integration with other systems such as review bots.

For example, one can utilize `jq` to process the JSON stdout into [the GitHub Actions annotation format](https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#example-creating-an-annotation-for-an-error).

```shell
iccheck --format json | jq -r '":::notice file=\(.filename),line=\(.start_l),endLine=\(.end_l),title=Possible missing change::Possible missing a consistent change here (L\(.start_l) - L\(.end_l), distance \(.distance))"'
```

### In GitHub Actions

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
