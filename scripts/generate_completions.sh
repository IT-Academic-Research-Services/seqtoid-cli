#!/bin/bash
set -euo pipefail

# Generate shell completions named by the COMMAND (seqtoid), not the shell, so that
# Homebrew/nfpm install them with idiomatic names: bash_completion.d/seqtoid, the zsh
# autoload file _seqtoid, and vendor_completions.d/seqtoid.fish. Homebrew derives the
# install target from the file's basename, so each shell gets its own subdirectory
# (the basename cannot be flat -- it would collide with the `seqtoid` binary at the root).
bin="$1"

mkdir -p completions/bash completions/zsh completions/fish completions/powershell
"$bin" completion bash       > completions/bash/seqtoid
"$bin" completion zsh        > completions/zsh/seqtoid
"$bin" completion fish       > completions/fish/seqtoid
"$bin" completion powershell > completions/powershell/seqtoid
