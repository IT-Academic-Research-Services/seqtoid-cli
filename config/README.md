# seqtoid-cli environment profiles

Ready-to-use config profiles for pointing the CLI at a SeqToID environment. These override
the binary's ldflag-baked defaults at runtime (viper), so **one package works against any env**.

| Env | base URL | Auth0 (alpha) |
|---|---|---|
| **dev** (`dev.yaml`) | `https://dev.seqtoid.org` | `auth.dev.seqtoid.org` |
| **env-staging** (`env-staging.yaml`) | `https://env-staging.seqtoid.org` | `auth.dev.seqtoid.org` (shared w/ dev during alpha) |

All values are public OAuth client config (native app / device flow — **no client secret**).

## Selecting an environment (three ways)

1. **`--config` flag** (explicit, per-invocation):
   ```
   seqtoid --config /path/to/config/env-staging.yaml metagenomics upload ...
   ```
2. **Default config file** — copy a profile to the OS config dir:
   ```
   cp config/env-staging.yaml ~/.config/seqtoid-cli/config.yaml   # Linux/macOS
   ```
3. **Environment variables** (viper prefix `SEQTOID_CLI_`), e.g.:
   ```
   export SEQTOID_CLI_SEQTOID_BASE_URL=https://env-staging.seqtoid.org
   export SEQTOID_CLI_AUTH0_HOST=auth.dev.seqtoid.org
   export SEQTOID_CLI_AUTH0_CLIENT_ID=ZUgpnLfFY64pS1LdW9PR6bMZFwhATuec
   export SEQTOID_CLI_AUTH0_AUDIENCE=https://seqtoid-dev.us.auth0.com/api/v2/
   ```

## Relationship to packaging (brew / Windows / Linux)

`.goreleaser.yml` bakes a single default env via ldflags (`AUTH0_CLIENT_ID`, `AUTH0_HOST`,
`AUTH0_AUDIENCE`, `SEQTOID_BASE_URL`). The **same** per-platform artifact (Homebrew cask,
Windows `.zip`, Linux `.deb`/`.rpm`/`.tar.gz`) can then target dev **or** env-staging by
applying one of these profiles — no separate per-env build required. Recommended: bake the
most common default (e.g. dev) and ship these profiles for the others.

## ⚠️ Post-alpha TODO

env-staging currently borrows dev's Auth0 tenant for alpha. When it reverts to its own staging
tenant, update `auth0_host` / `auth0_client_id` / `auth0_audience` in `env-staging.yaml`
(base URL is unchanged). Source of truth: the env's chamber `AUTH0_CLI_*` params.
