#!/usr/bin/env bash

set -euo pipefail

DATA_BASE_DIR=./repos/

<./0_github-ranking-2024-08-07.json \
jq -c \
'.[] | select(.item != "top-100-stars" and .item != "top-100-forks") | select(.rank | tonumber <= 3)' | \
while read -r repo_json; do
  echo "==> Next experiment: $repo_json"
  REPO_URL=$(jq -r '.repo_url' <<< "$repo_json")
  REPO_PATH=$(sed -e 's/https:\/\/github\.com\///' <<< "$REPO_URL")
  REPO_FULL_PATH="$DATA_BASE_DIR$REPO_PATH"
  if ! test -d "$REPO_FULL_PATH/.git"; then
    echo "Cloning repository ..."
    mkdir -p "$REPO_FULL_PATH"
    git clone "$REPO_URL" "$REPO_FULL_PATH"

    # Reset to old HEADs at the time of the original evaluation,
    # for a complete replication of the results
    COMMIT_HASH=$(grep "^${REPO_PATH} " 1_head.txt | awk '{print $2}')
    if [ -n "$COMMIT_HASH" ]; then
      git -C "$REPO_FULL_PATH" reset --hard "$COMMIT_HASH"
    fi
  fi
done
