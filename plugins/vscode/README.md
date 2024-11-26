# ICCheck Extension for VSCode

ICCheck finds inconsistent changes in your copy-pasted codes!

ICCheck takes any 2 revisions in a Git repository, and reports possible inconsistent changes.
Compare between HEAD and worktree for pre-commit changes, or between main and your feature branch
to run last-second checks on Pull Requests.

Please note that the way of detecting copy-pasted codes (i.e. cloned codes) is heuristic,
and therefore is not perfect.
While ICCheck detects many cloned codes correctly, it may also detect many false-positives.

This plugin was developed mainly for academic research.

Repository: https://github.com/salab/iccheck

<!--
Extension template code is copied and modified from
https://github.com/microsoft/vscode-extension-samples/tree/main/lsp-sample.
-->
