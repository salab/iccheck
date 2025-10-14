# Evaluation using the Dataset in CBCD paper

Executing files in filename order should replicate the results in the paper.

## Preparing repositories

Run `mkdir repos` and clone these repositories:

- git (https://github.com/git/git)
- linux (https://github.com/torvalds/linux)
- postgres (https://github.com/postgres/postgres)

## Run the experiment

`./1_run.sh` will (re)run the experiment.

The following scripts will use stdout of this script, so record the output to a temporary file:

```shell
./run.sh > 1_stdout.log 2> 1_stderr.log
```

## Analyzing the results

As stated in the paper, out of 53 cases, 15 cases are excluded from analysis.
This is because the query and the correct code snippet are not in the same snapshot,
making them unsuitable for evaluation of our iccheck tool.

Running `python3 2_cbcd_calc.py` will output the analysis results to stdout.
`python3 3_plot.py` will show the results in a graphical format.
