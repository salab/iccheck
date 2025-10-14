#!/usr/bin/env bash

set -euxo pipefail

DATA_BASE_DIR=./repos/
LOGS_DIR=./logs/

<./0_github-ranking-2024-08-07.json \
jq -c \
'.[] | select(.item != "top-100-stars" and .item != "top-100-forks") | select(.rank | tonumber <= 3)' | \
while read -r repo_json; do
  echo "==> Next experiment: $repo_json"
  REPO_URL=$(jq -r '.repo_url' <<< "$repo_json")
  REPO_PATH=$(sed -e 's/https:\/\/github\.com\///' <<< "$REPO_URL")
  REPO_FULL_PATH="$DATA_BASE_DIR$REPO_PATH"
  if ! test -d "$REPO_FULL_PATH/.git"; then
    echo "Error: repository not found, skipping"
    continue
  fi

  # Run experiment for repository
  REPO_LOG_PATH="$LOGS_DIR$REPO_PATH"
  mkdir -p "$REPO_LOG_PATH"
  set +e # Do not terminate the experiment even if one repo fails
  ./2_1_run-for-repo.sh "$REPO_FULL_PATH" > "$REPO_LOG_PATH"/stdout.log 2> "$REPO_LOG_PATH"/stderr.log
  set -e
done
