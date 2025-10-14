#!/usr/bin/env bash

set -euxo pipefail

shuf 5_2_clone_sets.ndjson | head -n 1000 > 5_4_shuffled.ndjson
