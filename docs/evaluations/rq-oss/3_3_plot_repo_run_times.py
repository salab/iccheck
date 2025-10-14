import json
import matplotlib.pyplot as plt
import numpy as np
import seaborn as sns
import matplotlib_fontja

import matplotlib
matplotlib.rcParams['pdf.fonttype'] = 42
matplotlib.rcParams['ps.fonttype'] = 42

if __name__ == '__main__':
    data = []
    with open("3_2_filtered.log") as f:
        for line in f:
            datum = json.loads(line)
            if "error" in datum["result"]:
                continue
            data.append(float(datum["time"]["real"]))

    print(f"mean: {np.mean(data)}, median: {np.median(data)}")

    plt.figure(figsize=(8, 2))
    ax = sns.boxplot(data=data, orient='h', color='gray')
    ax.set_xlim(left=0, right=10)

    plt.xlabel('Time (s)')
    # plt.title(f'Execution time of ICCheck (n={len(data)})')
    plt.tight_layout()

    # plt.savefig("3_4_repo-run-times.png")
    plt.savefig("3_4_repo-run-times.pdf")
    # plt.show()
