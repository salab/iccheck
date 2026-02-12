# Evaluations

This directory contains replication scripts and documentation for preliminary evaluations used in the Science of
Computer Programming journal entry.

## Preparation

- The documents and scripts assume that `iccheck` to be on the PATH. So, either:
    - Build it from the source and revision you want,
    - Or, download it from [the GitHub release](https://github.com/salab/iccheck/releases) assets.

## Data sources

Some files included in this directory are copied from the following sources:

- [./rq-cbcd/0_cbcd-dataset.json](./rq-cbcd/0_cbcd-dataset.json)
    - From [takashi-ishio/NCDSearch](https://github.com/takashi-ishio/NCDSearch/blob/main/evaluation-dataset/cbcd-dataset.json)
         - From [CBCD: Cloned buggy code detector | IEEE Conference Publication | IEEE Xplore](https://ieeexplore.ieee.org/abstract/document/6227183)
- [./rq-oss/0_github-ranking-2024-08-07.csv](./rq-oss/0_github-ranking-2024-08-07.csv)
    - Frtom [EvanLi/Github-Ranking](https://github.com/EvanLi/Github-Ranking/blob/97bf7eb092678fd2ac06411ba9316d3ff73f9b8c/Data/github-ranking-2024-08-07.csv)
- [./rq-oss/0_github-ranking-2024-08-07.json](./rq-oss/0_github-ranking-2024-08-07.json)
    - Converted with:
      `<0_github-ranking-2024-08-07.csv python3 -c 'import csv, json, sys; print(json.dumps([dict(r) for r in csv.DictReader(sys.stdin)]))'`
