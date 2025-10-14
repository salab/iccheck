import json
import matplotlib.pyplot as plt
import numpy as np
import seaborn as sns
import matplotlib_fontja

import matplotlib
matplotlib.rcParams['pdf.fonttype'] = 42
matplotlib.rcParams['ps.fonttype'] = 42

def missing_changes_cnt(result):
    return sum(len(clone_set["missing"]) for clone_set in result)

if __name__ == '__main__':
    data = []
    with open("3_2_filtered.log") as f:
        for line in f:
            datum = json.loads(line)
            if "error" in datum["result"]:
                continue
            data.append(float(missing_changes_cnt(datum["result"])))

    print(f"len: {len(data)}")
    data = [d for d in data if d > 0]
    print("filtered to at least one missing changes")
    print(f"len: {len(data)}")
    print(f"mean: {np.mean(data)}, median: {np.median(data)}, max: {np.max(data)}, sum: {np.sum(data)}")

    plt.figure(figsize=(8, 2))
    ax = sns.boxplot(data=data, orient='h', color='gray')
    ax.set_xlim(left=0, right=100)

    # plt.xlabel('Detected missing changes')

    plt.savefig("4_2_detected_missing_changes.pdf")
    # plt.show()
