import json
import subprocess
import re

repos = [
    "papers-we-love/papers-we-love",
    "TerryCavanagh/VVVVVV",
    "qinwf/awesome-R",
    "FelisCatus/SwitchyOmega",
    "tensorflow/tensorflow",
    "elixir-lang/elixir",
    "junegunn/vim-plug",
    "bregman-arie/devops-resources",
    "ryanoasis/nerd-fonts",
    "avelino/awesome-go",
    "terryum/awesome-deep-learning-papers",
    "Alamofire/Alamofire",
    "metabase/metabase",
    "ScoopInstaller/Scoop",
    "square/okhttp",
    "animate-css/animate.css",
    "rails/rails",
    "JuliaLang/julia",
    "microsoft/ML-For-Beginners",
    "microsoft/PowerToys",
    "freeCodeCamp/freeCodeCamp",
    "flutter/flutter",
    "jgm/pandoc",
    "torvalds/linux",
    "kubernetes/kubernetes",
    "iina/iina",
    "tauri-apps/tauri",
    "FluxML/Flux.jl",
    "danielmiessler/SecLists",
    "facebook/react",
    "facebook/react-native",
    "vim/vim",
    "GrowingGit/GitHub-Chinese-Top-Charts",
    "twbs/bootstrap",
    "phoenixframework/phoenix",
    "NvChad/NvChad",
    "ShiArthur03/ShiArthur03",
    "necolas/normalize.css",
    "tidyverse/ggplot2",
    "ChrisTitusTech/winutil",
    "prisma/prisma1",
    "android/architecture-samples",
    "vinta/awesome-python",
    "Gamua/Starling-Framework",
    "apache/spark",
    "exacity/deeplearningbook-chinese",
    "tonsky/FiraCode",
    "f/awesome-chatgpt-prompts",
    "netdata/netdata",
    "lib-pku/libpku",
    "brendangregg/FlameGraph",
    "luanfujun/deep-photo-styletransfer",
    "Kong/kong",
    "cxli233/FriendsDontLetFriends",
    "trekhleb/javascript-algorithms",
    "Solido/awesome-flutter",
    "rundeck/rundeck",
    "gradle/gradle",
    "so-fancy/diff-so-fancy",
    "Genymobile/scrcpy",
    "nagadomi/waifu2x",
    "AppFlowy-IO/AppFlowy",
    "logseq/logseq",
    "SDWebImage/SDWebImage",
    "Baystation12/Baystation12",
    "rust-lang/rust",
    "denoland/deno",
    "ohmyzsh/ohmyzsh",
    "vuejs/vue",
    "dotnet/core",
    "2dust/v2rayN",
    "jashkenas/coffeescript",
    "Snailclimb/JavaGuide",
    "mojs/mojs",
    "tgstation/tgstation",
    "public-apis/public-apis",
    "fonsp/Pluto.jl",
    "open-source-flash/open-source-flash",
    "electron/electron",
    "PostgREST/postgrest",
    "AFNetworking/AFNetworking",
    "ParadiseSS13/Paradise",
    "koalaman/shellcheck",
    "JetBrains/kotlin",
    "AlDanial/cloc",
    "TadasBaltrusaitis/OpenFace",
    "plausible/analytics",
    "d3/d3",
    "kamranahmedse/developer-roadmap",
    "neovim/neovim",
    "twitter/the-algorithm",
    "shadowsocks/shadowsocks-windows",
    "laravel/framework",
    "laravel/laravel",
    "krahets/hello-algo",
    "MustangYM/WeChatExtension-ForMac",
    "golang/go",
    "ripienaar/free-for-dev",
    "vsouza/awesome-ios",
    "jekyll/jekyll",
    "donnemartin/system-design-primer",
    "mastodon/mastodon",
]

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

ext_to_lang = {}
for lang in lang_extensions:
    for ext in lang_extensions[lang]:
        ext_to_lang[ext] = lang

def get_file_ext(filename):
    # strip directory name
    filename = re.sub(r'.*/', '', filename)
    # get file extension
    return re.sub(r'.*\.', '.', filename)

def to_file_url(repo, commit, file):
    filename = file["filename"]
    start = file["start_l"]
    end = file["end_l"]
    line_hash = f"L{start}" if start == end else f"L{start}-L{end}"
    return f"https://github.com/{repo}/blob/{commit}/{filename}#{line_hash}"

if __name__ == '__main__':
    repo_map = {}
    for repo in repos:
        parts = repo.split("/")
        if parts[1] in repo_map:
            raise "repo name not unique!"
        repo_map[parts[1]] = repo

    clone_sets = []
    with open("3_2_filtered.log") as f:
        for line in f:
            datum = json.loads(line)
            if "error" in datum["result"]:
                continue
            if len(datum["result"]) == 0:
                continue

            exp = datum["experiment"]
            repo = repo_map[exp["dir"]]
            commit = exp["commit"]
            time = datum["time"]["real"]

            res = subprocess.run(["git", "rev-parse", commit], cwd=f"./repos/{repo}", stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            if res.returncode != 0:
                raise f"git rev-parse failed: {res.stdout.strip()} {res.stderr.strip()}"
            full_commit = res.stdout.strip()

            for cs in datum["result"]:
                exts = {get_file_ext(f["filename"]) for f in cs["changed"]} | {get_file_ext(f["filename"]) for f in cs["missing"]}
                langs = {ext_to_lang.get(ext, "") for ext in exts} - {""}
                clone_sets.append({
                    "repo": repo,
                    "commit": commit,
                    "url": f"https://github.com/{repo}/commit/{full_commit}",
                    "time": time,
                    "cmd": f"iccheck -r ./repos/{repo} --from {commit}^ --to {commit}",
                    "summary": f"{len(cs["changed"])} chunks changed, {len(cs["missing"])} chunks missing change",
                    "changed": ", ".join([to_file_url(repo, full_commit, f) for f in cs["changed"]]),
                    "missing": ", ".join([to_file_url(repo, full_commit, f) for f in cs["missing"]]),
                    "exts": ", ".join(sorted(list(exts))),
                    "langs": ",".join(sorted(list(langs))),
                })

    print("clone set size", len(clone_sets))

    with open("5_2_clone_sets.ndjson", "wt+") as f:
        for cs in clone_sets:
            line = json.dumps(cs, separators=(',', ':'))
            f.write(line)
            f.write("\n")
