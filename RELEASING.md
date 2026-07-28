# Releasing the SeqToID CLI

The `seqtoid` binary is built and published by [GoReleaser](https://goreleaser.com/) from
[`.github/workflows/cicd.yml`](.github/workflows/cicd.yml). Cross-platform binaries, checksums,
Linux packages (`.deb`/`.rpm`), and a Homebrew cask are produced automatically when a GitHub
Release is created.

## How a release fires

The `Build`, `Test`, `Dependencies`, and `Lint` jobs run on every push. The `Prerelease` and
`Release` jobs only run on a **GitHub Release** event:

- **Prerelease** — create a GitHub Release with *"Set as a pre-release"* checked. Runs
  `goreleaser release --config .goreleaser.prerelease.yml` (binaries + checksums only, no Homebrew).
- **Release** — create a normal (non-prerelease) GitHub Release. Runs
  `goreleaser release --config .goreleaser.yml` (binaries + checksums + `.deb`/`.rpm` + Homebrew cask).

Steps to cut a production release:

1. Ensure `main` is green.
2. Tag with a semver tag, e.g. `git tag v6.1.0 && git push origin v6.1.0`.
3. Create a GitHub Release from that tag (GitHub UI or `gh release create v6.1.0 --generate-notes`).
4. The `Release` job publishes the artifacts to the Release and updates the Homebrew tap.

The version string baked into the binary (`seqtoid version`, and the `client` sent on upload — the
server enforces `MIN_CLI_VERSION`) comes from the git tag via the `pkg.Version` ldflag.

## Required repository secrets

The release bakes environment config into the binary at build time via ldflags. Set these as
**repository secrets** (Settings -> Secrets and variables -> Actions). For the public production
CLI these are the **production** SeqToID / Auth0 values:

| Secret | Purpose | Example (prod) |
| --- | --- | --- |
| `AUTH0_CLIENT_ID` | Auth0 **Native app** client id for the CLI (device flow). Also the id_token `aud` the server's `verify_cli` checks. | (prod CLI app client id) |
| `AUTH0_HOST` | Auth0 domain the CLI talks to (custom domain). | `auth.seqtoid.org` |
| `AUTH0_AUDIENCE` | Audience requested during the device flow so Auth0 issues a token. Use a valid API identifier for the tenant (e.g. the Management API). | `https://<tenant>.us.auth0.com/api/v2/` |
| `SEQTOID_BASE_URL` | The SeqToID web API the CLI uploads to. | `https://seqtoid.org` |
| `HOMEBREW_TAP_GITHUB_TOKEN` | A token with **write** access to the Homebrew tap repo (below). Release job only. | (bot PAT) |

`GITHUB_TOKEN` is provided automatically by Actions.

> One release targets **one** environment (whatever the secrets point at — production for the public
> build). To use the CLI against dev/staging, build locally or pass `--config <file>` with
> `auth0_client_id` / `auth0_host` / `auth0_audience` / `seqtoid_base_url` overrides; the baked
> defaults are only the fallback.

## What must be set up on the UCSF / org-owner side

These need `IT-Academic-Research-Services` **org-owner** rights (a repo-member identity cannot do
them):

1. **Homebrew tap repo** — create `IT-Academic-Research-Services/homebrew-tap` (public, so users can
   `brew install IT-Academic-Research-Services/tap/seqtoid`). The `Release` job pushes the generated
   cask to it.
2. **`HOMEBREW_TAP_GITHUB_TOKEN`** — a bot/service token with write access to that tap repo, stored
   as a secret on this repo.
3. **Distribution visibility** — this repo is currently **private**, so GitHub Release assets are not
   publicly downloadable. To distribute to end users, either make `seqtoid-cli` public or host the
   release assets on a public channel. Decide with UCSF.
4. **Production Auth0 CLI app** — a Native app (Device Code grant enabled) in the prod tenant;
   its client id becomes `AUTH0_CLIENT_ID`.

## Not yet done (follow-ups)

- **Code signing / notarization** — the macOS binaries are not signed, so Gatekeeper would quarantine
  them. The cask strips the quarantine attribute on install as an interim measure; proper signing +
  notarization (and Windows signing) is a follow-up.
- The first real release should be validated end to end (download each OS artifact, `brew install`
  from the tap, confirm `seqtoid version` and an upload against the target environment).

## Local build

`make build` produces a `seqtoid` binary using the same ldflags, reading `AUTH0_CLIENT_ID`,
`AUTH0_HOST`, `AUTH0_AUDIENCE`, `SEQTOID_BASE_URL`, and `VERSION` from the environment. To dry-run the
full release build without publishing: `goreleaser release --snapshot --clean`.
