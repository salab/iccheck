#!/usr/bin/env bash

set -euo pipefail

run-experiment() {
  if [ "$path" = "$lookup_path" ]; then
    lookup_eq=true
  else
    lookup_eq=false
  fi

  EXPERIMENT=$(jq -n "{repo: \"$repo\", commit: \"$commit\", file: \"$file\", sline: \"$sline\", eline: \"$eline\", lookup_eq: \"$lookup_eq\"}")
  >&2 echo "=====> Experiment $repo $commit $file $sline $eline (query = lookup?: $lookup_eq)"
  >&2 pushd "./repos/$repo"

  # Run iccheck
  >&2 echo "==> Detecting inconsistencies"
  TIME_TMP=$(mktemp)
  set +e
  RAW_OUT=$(command time -o "$TIME_TMP" --quiet --format '{"real": %e, "user": %U, "system": %S}' iccheck search --ref "$commit" --file "$file" --start-line "$sline" --end-line "$eline" --format json --timeout-seconds 300)
  STATUS=$?
  set -e

  # Echo json result to stdout
  if [ $STATUS -eq 0 ]; then
    RESULT=$(jq -s <<< "$RAW_OUT")
  else
    RESULT="{\"error\": $STATUS}"
  fi
  echo "$EXPERIMENT" "$(cat "$TIME_TMP")" "$RESULT" | jq -s -c '{experiment: .[0], time: .[1], result: .[2]}' | tee >(cat >&2)
  rm "$TIME_TMP"

  # Log separator
  >&2 popd
  >&2 echo ""
  >&2 echo ""
  >&2 echo ""
}

< ./0_cbcd-dataset.json jq -r '.queries | values[] | "\(.path) \(.queryloc.path) \(.queryloc.file) \(.queryloc.sline) \(.queryloc.eline)"' | \
#< ./0_cbcd-dataset.json jq -r '.queries."16" | "\(.path) \(.queryloc.path) \(.queryloc.file) \(.queryloc.sline) \(.queryloc.eline)"' | \
while read -r lookup_path path file sline eline; do
  IFS="-" read -r repo commit <<< "$path"
  run-experiment
done
