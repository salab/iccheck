#!/usr/bin/env bash

set -eu

EXPERIMENT_DEPTH=100

if [ "$#" -ne 1 ]; then
  >&2 echo "Usage: $0 [git-directory]"
  exit 1
fi

cd "$1"

DIR=$(basename "$(pwd)")
DEFAULT_BRANCH=$(git remote show origin | sed -n '/HEAD branch/s/.*: //p')

>&2 git reset --hard HEAD
>&2 git checkout -f "$DEFAULT_BRANCH"

git log --oneline --first-parent --pretty=format:"%h" | \
# This line causes SIGPIPE
head -n "$EXPERIMENT_DEPTH" | \
while read -r commit; do
  >&2 echo "==> Next experiment: $commit"
  EXPERIMENT=$(jq -n "{dir: \"$DIR\", commit: \"$commit\"}")

  # Run iccheck
  TIME_TMP=$(mktemp)
  set +e
  RAW_OUT=$(command time -o "$TIME_TMP" --quiet --format '{"real": %e, "user": %U, "system": %S}' iccheck --from "$commit^" --to "$commit" --format json --timeout-seconds 60)
  STATUS=$?
  set -e

  # Echo result (in a machine-readable manner; json) to stdout
  if [ $STATUS -eq 0 ]; then
    RESULT=$(jq -s <<< "$RAW_OUT")
  else
    RESULT="{\"error\": $STATUS}"
  fi
  echo "$EXPERIMENT" "$(cat "$TIME_TMP")" "$RESULT" | jq -s -c '{experiment: .[0], time: .[1], result: .[2]}' | tee >(cat >&2)
  rm "$TIME_TMP"
done
