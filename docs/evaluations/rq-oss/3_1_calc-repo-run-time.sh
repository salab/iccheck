#!/usr/bin/env bash

set -euo pipefail

echo "Number of instances:"
find ./logs/ -name stdout.log -print0 | xargs -0 cat | \
wc -l

filter_commits_by_files_changed() {
    local repo_dir=$1
    local max_files=$2
    shift
    while IFS= read -r line; do
      commit=$(<<<"$line" jq -r '.experiment.commit')
      files_changed=$(git -C "$repo_dir" diff --name-only "$commit^" "$commit" | wc -l)
      if [ "$files_changed" -le "$max_files" ]; then
        echo "$line"
      fi
    done
}

echo "Number of instances with <= 25 files changed:"
find ./logs/ -name stdout.log | while read -r logfile; do
  repo="${logfile#./logs/}"
  repo=$(dirname "$repo")
  <"$logfile" \
    filter_commits_by_files_changed "./repos/$repo" 25
done > 3_2_filtered.log
<3_2_filtered.log wc -l

echo "Out of which, success instances:"
<3_2_filtered.log \
  jq 'select(.result | type == "array")' -c | \
  wc -l
