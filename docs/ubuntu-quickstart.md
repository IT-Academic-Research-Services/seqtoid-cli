# SeqToID CLI -- Ubuntu / Linux setup (env-staging)

Install the `seqtoid` command-line tool on Ubuntu (or any Linux) and point it at
**env-staging**. Current release: **v6.1.4**.

Linux ships as a `.deb` package (Debian/Ubuntu), an `.rpm` package (Fedora/RHEL), or a
plain tarball -- for both **amd64** (Intel/AMD) and **arm64** (Graviton/Ampere/Raspberry Pi
and other ARM). Pick the file that matches your machine.

## Prerequisites

- **A SeqToID account on env-staging** -- log in once at <https://env-staging.seqtoid.org>
  in a browser first (Google/email). The CLI uses the same identity, and a first-time
  account is created automatically.

## 1. Pick your architecture

```bash
dpkg --print-architecture   # Debian/Ubuntu: prints "amd64" or "arm64"
uname -m                     # any Linux: "x86_64" -> amd64, "aarch64" -> arm64
```

Download the matching asset from the
[latest release](https://github.com/IT-Academic-Research-Services/seqtoid-cli/releases/latest):

| Your machine | `.deb` file | `.rpm` file | Tarball |
|---|---|---|---|
| amd64 / x86_64 | `seqtoid-cli_linux_amd64.deb` | `seqtoid-cli_linux_amd64.rpm` | `seqtoid-cli_linux_amd64.tar.gz` |
| arm64 / aarch64 | `seqtoid-cli_linux_arm64.deb` | `seqtoid-cli_linux_arm64.rpm` | `seqtoid-cli_linux_arm64.tar.gz` |

## 2. Install

### Option A -- `.deb` (Debian / Ubuntu, recommended)

```bash
# amd64:
sudo dpkg -i seqtoid-cli_linux_amd64.deb
# or arm64:
sudo dpkg -i seqtoid-cli_linux_arm64.deb
```

This puts `seqtoid` on your `PATH` at `/usr/bin/seqtoid` and installs shell completions
(bash under `/etc/bash_completion.d/`, zsh and fish under `/usr/share/`) plus the
per-environment config profiles under `/etc/seqtoid-cli/`.

### Option B -- `.rpm` (Fedora / RHEL / openSUSE)

```bash
sudo rpm -i seqtoid-cli_linux_amd64.rpm    # or the _arm64.rpm
```

### Option C -- tarball (no package manager, or unusual distro)

```bash
tar -xzf seqtoid-cli_linux_amd64.tar.gz    # or the _arm64.tar.gz
# put the binary on your PATH, e.g. a user-local bin dir:
mkdir -p ~/.local/bin && mv seqtoid ~/.local/bin/
# ensure ~/.local/bin is on PATH (most Ubuntu setups add it automatically):
command -v seqtoid || echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
```

Verify the install:

```bash
seqtoid version    # should print 6.1.4
```

## 3. Point the CLI at env-staging

The CLI defaults to **dev**. Point it at **env-staging** by setting one environment
variable. Paste this block into your terminal -- it detects the shell you're actually in,
sets the variable for the current session, and appends it to the right startup file so it
persists. It's safe to run more than once (it won't add a duplicate line):

```bash
export SEQTOID_CLI_SEQTOID_BASE_URL=https://env-staging.seqtoid.org
if [ -n "$ZSH_VERSION" ]; then RC=~/.zshrc
elif [ -n "$BASH_VERSION" ]; then RC=~/.bashrc
else RC=~/.profile; fi
grep -qF 'SEQTOID_CLI_SEQTOID_BASE_URL' "$RC" 2>/dev/null \
  || echo 'export SEQTOID_CLI_SEQTOID_BASE_URL=https://env-staging.seqtoid.org' >> "$RC"
```

> On **Linux**, interactive **bash** reads `~/.bashrc` (this is different from macOS, where
> interactive bash reads `~/.bash_profile`). The snippet above already picks `~/.bashrc` for
> bash and `~/.zshrc` for zsh, so you don't have to think about it.

Prefer to do it by hand? Add this line to your shell's startup file -- `~/.bashrc` for
**bash**, `~/.zshrc` for **zsh** -- then open a new terminal (or `source` that file):

```bash
export SEQTOID_CLI_SEQTOID_BASE_URL=https://env-staging.seqtoid.org
```

Alternatively, use a shipped config profile instead of the env var. The `.deb`/`.rpm`
install the profiles to `/etc/seqtoid-cli/`, and the tarball ships them under `config/`:

```bash
# per-command:
seqtoid --config /etc/seqtoid-cli/env-staging.yaml <command>

# or make it the default (Linux config dir is ~/.config/seqtoid-cli):
mkdir -p ~/.config/seqtoid-cli
cp /etc/seqtoid-cli/env-staging.yaml ~/.config/seqtoid-cli/config.yaml
```

> Auth is already baked correctly for the alpha (shared dev Auth0 tenant,
> `auth.dev.seqtoid.org`) -- you only ever override the **base URL** to switch environments.

## 4. Log in

```bash
seqtoid login
```

This opens a browser device-flow page. Approve the code shown and sign in.
(During alpha the login page is `auth.dev.seqtoid.org`; that's expected -- env-staging
shares the dev Auth0 tenant.)

## 5. Accept the user agreement

```bash
seqtoid accept-user-agreement
```

## 6. Verify + first upload

```bash
# sanity check you are hitting env-staging and authenticated
seqtoid list-metadata-for-host-organism Human
```

The **project must already exist** -- create it in the env-staging web UI first.
Then upload a paired-end Illumina sample:

```bash
seqtoid metagenomics upload-sample sample_R1.fastq.gz sample_R2.fastq.gz \
  --project "Your Project Name" \
  --sequencing-platform Illumina \
  --metadata-csv metadata.csv
```

Your samples appear under your account at <https://env-staging.seqtoid.org>.

### Metadata

`--metadata-csv` is the reliable path. Generate a template, fill it in, and pass it:

```bash
seqtoid generate-metadata-template for-sample-name "sample_R1" -o metadata.csv
# edit metadata.csv, then upload as above
```

Inline metadata also works (commas in values, e.g. locations, are handled as of v6.1.4):

```bash
  -m "Host Organism=Human" -m "Collection Location=Santa Barbara, CA, USA"
```

---

## Quick reference (in order)

1. `dpkg --print-architecture` (or `uname -m`) to pick amd64 vs arm64
2. `sudo dpkg -i seqtoid-cli_linux_<arch>.deb`  (or `.rpm` / tarball)
3. add the `SEQTOID_CLI_SEQTOID_BASE_URL` line for env-staging (snippet in step 3), then open a new terminal
4. `seqtoid login`  (approve in browser)
5. `seqtoid accept-user-agreement`
6. `seqtoid metagenomics upload-sample ...`

## Troubleshooting

| Symptom | Fix |
|---|---|
| Uploaded but can't find samples in env-staging | The CLI is pointed at dev. Check `echo $SEQTOID_CLI_SEQTOID_BASE_URL` (or your `--config`). |
| `seqtoid: command not found` after tarball install | `seqtoid` isn't on your `PATH`. Move it to `~/.local/bin` (see step 2, Option C) and open a new terminal. |
| `dpkg: error ... wrong architecture` | You grabbed the wrong file. Re-check `dpkg --print-architecture` and download the matching `_amd64`/`_arm64` package. |
| Tab-completion not working | Open a new shell after install; for bash ensure the `bash-completion` package is installed (`sudo apt install bash-completion`). |
| `Please upgrade your CLI` / HTTP 426 | Your CLI is older than the server minimum (`6.0.0`). Download and install the latest release. |
| `Cannot upload ... accept the user agreement` | Run `seqtoid accept-user-agreement` (step 5). |
| Login page shows `auth.dev.seqtoid.org` | Expected during alpha. |

For the full profile/override reference see [`config/README.md`](../config/README.md).
