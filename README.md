# ICCheck - Inconsistent Change Checker

A work-in-progress tool which finds inconsistent changes in your (pre-)commits / Pull Requests.

## Usage

To be filled

```text
Finds inconsistent changes in your git changes

Usage:
  iccheck [flags]

Flags:
  -f, --from string        Target git ref to compare against. Usually earlier in time. (default "main")
  -h, --help               help for iccheck
      --log-level string   Log level (debug, info, warn, error) (default "info")
  -r, --repo string        Source git directory (default ".")
  -t, --to string          Source git ref to compare from. Usually later in time. (default "HEAD")
```

### In GitHub Actions

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
      - run: iccheck --from "$ICCHECK_FROM" --to "$ICCHECK_TO"
```
