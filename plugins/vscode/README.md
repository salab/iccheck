# ICCheck Extension for VSCode

ICCheck takes any 2 revisions (including commit and worktree) from a Git repository
and reports possible inconsistent changes in pre-commit files, commited files, and/or Pull Requests.

ICCheck plugin lists pre-commit changes made on a git repository
and checks missing changes on cloned codes (i.e. copy-pasted codes).

The plugin and the method for detecting cloned code is still in development and being improved.
While ICCheck detects many cloned codes correctly, note that it may also detect many false-positives.

This plugin was developed mainly for academic research.

Repository: https://github.com/salab/iccheck

<!--
Extension template code is copied and modified from
https://github.com/microsoft/vscode-extension-samples/tree/main/lsp-sample.
-->
