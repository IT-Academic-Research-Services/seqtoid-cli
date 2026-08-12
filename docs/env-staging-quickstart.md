# Alpha tester quickstart — using the CLI against env-staging

This is the end-to-end path for an alpha tester to install the `seqtoid` CLI, point it at
**env-staging** (`https://env-staging.seqtoid.org`), log in, and upload samples.

> **You need first:** a SeqToID account on env-staging. Just log in once at
> <https://env-staging.seqtoid.org> in a browser (Google/email). Your CLI login uses the
> same identity, and a first-time account is provisioned automatically.

---

## 1. Install

**macOS (Homebrew):**
```bash
brew install IT-Academic-Research-Services/tap/seqtoid
seqtoid version
```

**Linux / Windows (or no Homebrew):** download the archive for your platform from the
[latest release](https://github.com/IT-Academic-Research-Services/seqtoid-cli/releases/latest),
unpack it, and put `seqtoid` on your `PATH`. On Ubuntu/Linux there are also `.deb`/`.rpm`
packages (amd64 and arm64) -- see the [Ubuntu / Linux quickstart](./ubuntu-quickstart.md).

## 2. Point the CLI at env-staging

The CLI defaults to **dev**. To target **env-staging**, pick ONE of these:

**A. Environment variable (simplest — works with every install, no file needed):**
```bash
export SEQTOID_CLI_SEQTOID_BASE_URL=https://env-staging.seqtoid.org
```
Set this in your shell rc (`~/.zshrc` / `~/.bashrc`) so it persists. All CLI commands then hit env-staging.

**B. Config profile with `--config`** — the release archive ships `config/env-staging.yaml`; point `--config` at it:
```bash
# Tarball install: the profile is in the archive you unpacked
seqtoid --config ./config/env-staging.yaml <command>

# Homebrew install: the profile ships in the Caskroom (version is in the path)
seqtoid --config "$(brew --prefix)/Caskroom/seqtoid/$(seqtoid version)/config/env-staging.yaml" <command>
```

**C. Default config file (no flag each time)** — copy the profile to the CLI's config directory. That directory
is **OS-specific** (`os.UserConfigDir`), so use the right one:
```bash
# macOS:
mkdir -p "$HOME/Library/Application Support/seqtoid-cli"
cp <path-to>/env-staging.yaml "$HOME/Library/Application Support/seqtoid-cli/config.yaml"

# Linux:
mkdir -p ~/.config/seqtoid-cli
cp <path-to>/env-staging.yaml ~/.config/seqtoid-cli/config.yaml
```

> Prefer **A** (env var) unless you want a checked-in profile — it's identical on every platform and needs no
> path lookups. On macOS the config dir is `~/Library/Application Support/seqtoid-cli`, **not** `~/.config`.

> Auth is already baked correctly for the alpha (shared dev Auth0 tenant,
> `auth.dev.seqtoid.org`) — you only ever override the **base URL** to switch environments.

## 3. Log in

```bash
seqtoid login
```
This opens a browser device-flow page. Approve it, and the CLI caches your session.
(First time only, accept the user agreement: `seqtoid accept-user-agreement`.)

## 4. Verify + upload

```bash
# sanity check you are hitting env-staging and authenticated
seqtoid list-metadata-for-host-organism Human

# example metagenomics upload (see `seqtoid metagenomics upload --help`)
seqtoid metagenomics upload \
  --project "<your project>" \
  --metadata metadata.csv \
  sample_R1.fastq.gz sample_R2.fastq.gz
```

Your samples appear under your account at <https://env-staging.seqtoid.org>.

---

## Troubleshooting

| Symptom | Cause / fix |
|---|---|
| Uploaded but can't find samples in env-staging | The CLI is still pointed at **dev** (the default). Confirm step 2 — `echo $SEQTOID_CLI_SEQTOID_BASE_URL` or your `--config`. |
| `Please upgrade your CLI` / HTTP 426 | Your CLI is older than the server minimum (`6.0.0`). Reinstall the latest release. |
| Login browser page says `auth.dev.seqtoid.org` | Expected — env-staging shares the dev Auth0 tenant during alpha. |
| `Cannot upload ... accept the user agreement` | Run `seqtoid accept-user-agreement` (or `export SEQTOID_CLI_ACCEPTED_USER_AGREEMENT=Y`). |

For the full profile/override reference see [`config/README.md`](../config/README.md).
