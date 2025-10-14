import json
import re
import matplotlib.pyplot as plt
import numpy as np
import seaborn as sns
import matplotlib_fontja

import matplotlib
matplotlib.rcParams['pdf.fonttype'] = 42
matplotlib.rcParams['ps.fonttype'] = 42

lang_extensions = {
    "ActionScript": [".as"],
    "C": [".c"],
    "C#": [".cs"],
    "C++": [".cpp", ".cc", ".cxx", ".h", ".hpp"],
    "Clojure": [".clj", ".cljs", ".cljc"],
    "CoffeeScript": [".coffee"],
    "CSS": [".css"],
    "Dart": [".dart"],
    "DM": [".dm"],
    "Elixir": [".ex", ".exs"],
    "Go": [".go"],
    "Groovy": [".groovy"],
    "Haskell": [".hs"],
    "HTML": [".html", ".htm"],
    "Java": [".java"],
    "JavaScript": [".js"],
    "Julia": [".jl"],
    "Kotlin": [".kt", ".kts"],
    "Lua": [".lua"],
    "MATLAB": [".m"],
    "Objective-C": [".m", ".h"],
    "Perl": [".pl"],
    "PHP": [".php"],
    "PowerShell": [".ps1"],
    "Python": [".py"],
    "R": [".r", ".R"],
    "Ruby": [".rb"],
    "Rust": [".rs"],
    "Scala": [".scala"],
    "Shell": [".sh"],
    "Swift": [".swift"],
    "TeX": [".tex"],
    "TypeScript": [".ts"],
    "Vim script": [".vim"],
}

if __name__ == '__main__':
    ext_counts = {}
    with open("3_2_filtered.log") as f:
        for line in f:
            datum = json.loads(line)
            if "error" in datum["result"]:
                continue
            for clone_set in datum["result"]:
                for file in clone_set["missing"]:
                    filename = file["filename"]
                    # strip directory name
                    filename = re.sub(r'.*/', '', filename)
                    # get file extension
                    ext = re.sub(r'.*\.', '.', filename)
                    ext_counts[ext] = ext_counts.get(ext, 0) + 1

    ext_count_pairs = ext_counts.items()
    ext_count_pairs = sorted(ext_count_pairs, key=lambda x: -x[1])
    print("--- per extension:")
    for item in ext_count_pairs:
        print(f"{item[0]}: {item[1]}")

    categories = []
    values = []
    print("--- per language:")
    for lang, exts in lang_extensions.items():
        cnt = sum(ext_counts.get(ext, 0) for ext in exts)
        categories.append(lang)
        values.append(cnt)
        print(f"{lang}: {cnt}")

    # Sort
    sorted_indices = sorted(range(len(values)), key=lambda i: values[i], reverse=True)
    categories = [categories[i] for i in sorted_indices]
    values = [values[i] for i in sorted_indices]

    # Create a bar graph
    plt.figure(figsize=(7, 6))
    bars = plt.bar(categories, values, color='grey')

    # Add labels and title
    # plt.xlabel('Language')
    plt.xticks(rotation=80)
    plt.ylabel('Number of suggestions')
    plt.yscale('log')
    # plt.title('Number of suggestions per programming language')
    plt.tight_layout()

    plt.savefig("4_4_lang.pdf")
    # plt.show()
