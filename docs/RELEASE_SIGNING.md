# Release Artifact Signing — What We Need and How

Status: **nothing is signed yet.** This doc lists every artifact a release produces that *should* be
signed, what each one requires, and the concrete steps — split by who owns them (UCSF/procurement vs.
repo secrets vs. engineering/pipeline).

Priority order: **macOS first** (most user-visible via Gatekeeper), then Windows, then Linux packages
and checksums. None of this blocks distributing via Homebrew/Scoop today — it hardens the
direct-download experience and download integrity for GA.

---

## Current state / why this matters

| Platform | If unsigned, the user sees… | Covered today? |
| --- | --- | --- |
| macOS | Gatekeeper: *"cannot be opened because the developer cannot be verified"* | Partially — the Homebrew **cask strips the quarantine flag on install**, so `brew` users are fine. **Direct-download users are not.** |
| Windows | SmartScreen: *"Windows protected your PC"* / unknown publisher | No |
| Linux `.deb`/`.rpm` | `apt`/`dnf` warns or refuses unsigned packages | No |
| All (checksums) | No way to cryptographically verify the download is authentic | No |

---

## What gets signed (overview)

| Artifact | Mechanism | Credential needed | Where it runs |
| --- | --- | --- | --- |
| macOS binaries (arm64 + amd64) | Apple codesign + **notarization** | Apple **Developer ID Application** cert + App Store Connect API key | Linux runner via `rcodesign`, or a macOS runner |
| Windows `.exe` | **Authenticode** | Code-signing cert on HSM, **or** Azure Trusted Signing | Linux runner via `osslsigncode`/`jsign`, or Windows runner |
| Linux `.deb`/`.rpm` | **GPG** package signature | Project GPG key | Linux runner (GoReleaser `nfpms`) |
| `checksums.txt` (all) | **GPG** (or cosign) detached signature | Project GPG key (or cosign key) | Linux runner (GoReleaser `signs`) |

---

## 1. macOS binaries

**Goal:** signed + notarized so Gatekeeper accepts the binary for direct-download users.

### What we need (UCSF / Apple — the gating dependency)
1. **Apple Developer Program membership for the org** (~$99/yr, needs the org legal entity + D-U-N-S number).
   - ⚠️ **Check first whether UCSF already has an organizational Apple Developer team** — the cert must come from *their* Team ID, not a personal account.
2. **A "Developer ID Application" certificate** (the variant for distribution *outside* the Mac App Store), exported with its private key as a **`.p12`**.
3. **An App Store Connect API key** for notarization: an **Issuer ID**, **Key ID**, and the **`.p8`** private key file. (This is the modern, no-Apple-ID-password method.)

### Repo secrets to add
- `MACOS_CERT_P12_BASE64` (base64 of the `.p12`) + `MACOS_CERT_PASSWORD`
- `APPLE_API_ISSUER_ID`, `APPLE_API_KEY_ID`, `APPLE_API_KEY_P8_BASE64`

### Steps / pipeline wiring (engineering)
1. Build with **hardened runtime + secure timestamp** (required for notarization).
2. Sign the darwin binaries with the Developer ID Application identity.
3. Submit for **notarization** and wait for the ticket.
4. On the current Linux CI, use **`rcodesign`** (`apple-platform-rs`) — it can sign *and* notarize
   Mach-O binaries from Linux (no Mac needed). Alternative: a macOS runner with `codesign` +
   `notarytool`.

### Decision: bare binary vs `.pkg`
You **cannot staple** a notarization ticket to a bare executable — stapling only works on
`.pkg`/`.dmg`/`.app`.
- **Bare signed+notarized binary** → Gatekeeper accepts it **when the machine is online** (checks Apple). Simplest; fine for most researchers.
- **`.pkg` installer** (signed + notarized + **stapled**) → works fully **offline**; nicer for enterprise/managed Macs. Extra pipeline work.

Recommendation: start with the bare signed+notarized binary; add a `.pkg` only if offline/managed installs become a requirement.

---

## 2. Windows `.exe`

**Goal:** Authenticode-signed so SmartScreen doesn't flag "unknown publisher."

### What we need (UCSF / procurement) — pick ONE
- **OV (Organization Validation) code-signing certificate** — cheaper (~$200–400/yr), but SmartScreen reputation is **earned over time**, so early users still see warnings.
- **EV (Extended Validation) certificate** — **instant** SmartScreen reputation, but pricier and the key must live on a hardware token / cloud HSM.
- **Azure Trusted Signing** (Microsoft's managed service, ~$10/mo) — no hardware token, good SmartScreen standing, increasingly the recommended path. **Likely the best option for us.**

> ⚠️ Since June 2023, CA/Browser Forum rules require code-signing private keys to live on **FIPS
> 140-2 hardware** (HSM/token) — plain `.pfx` file signing is no longer issued for new certs. This is
> why a managed service (Azure Trusted Signing) or an HSM/token is required, not just a cert file.

### Repo secrets to add
- Azure Trusted Signing: an Azure service-principal credential (tenant/client/secret) + the signing account/profile names.
- (Or, for an HSM/token cert: the provider's credentials — varies by CA.)

### Steps / pipeline wiring (engineering)
- Sign the Windows `.exe` with **`signtool`** (Windows runner) or **`osslsigncode`/`jsign`** (Linux runner). Azure Trusted Signing has a dedicated signing tool/GitHub Action.

---

## 3. Linux `.deb` / `.rpm`

**Goal:** GPG-signed packages so `apt`/`dnf` trust them and users can verify.

### What we need
- A **project GPG key** (identity e.g. `Academic Research Services <its-arssupport@ucsf.edu>`). We can generate this ourselves — no external CA needed.

### Repo secrets to add
- `GPG_PRIVATE_KEY` (armored) + `GPG_PASSPHRASE`

### Steps / pipeline wiring (engineering)
1. GoReleaser `nfpms` signs the `.deb`/`.rpm` with the GPG key at build time.
2. **Publish the public key** (in the repo + on the releases page / a well-known URL) so users can
   import it to verify.
3. (Optional, larger scope) stand up a real apt/yum repo with signed metadata — only if we want
   `apt install seqtoid` rather than "download the `.deb` and verify."

---

## 4. Release checksums (all platforms)

**Goal:** anyone, on any OS, can verify a download is authentic and untampered.

### What we need
- The **same project GPG key** as §3 (or a **cosign** key for a Sigstore-style flow).

### Repo secrets to add
- Reuse `GPG_PRIVATE_KEY` / `GPG_PASSPHRASE` (or `COSIGN_KEY` / `COSIGN_PASSWORD`).

### Steps / pipeline wiring (engineering)
- GoReleaser `signs:` block produces a **detached signature of `checksums.txt`** on every release.
- Publish the public key alongside releases with verification instructions.

This is the cheapest, highest-leverage item — one GPG key covers both §3 and §4 and gives every user
a verification path even before per-platform signing is in place.

---

## Consolidated checklist

### UCSF / procurement (external — the long-lead items)
- [ ] Confirm whether UCSF already has an **organizational Apple Developer** team; if not, enroll.
- [ ] Issue a **Developer ID Application** certificate → export as `.p12`.
- [ ] Create an **App Store Connect API key** (Issuer ID + Key ID + `.p8`).
- [ ] Decide the **Windows** signing route (recommend **Azure Trusted Signing**) and provision it.

### Repo secrets (once the above exist)
- [ ] macOS: `MACOS_CERT_P12_BASE64`, `MACOS_CERT_PASSWORD`, `APPLE_API_ISSUER_ID`, `APPLE_API_KEY_ID`, `APPLE_API_KEY_P8_BASE64`
- [ ] Windows: Azure Trusted Signing service-principal creds (or HSM/token creds)
- [ ] Linux + checksums: `GPG_PRIVATE_KEY`, `GPG_PASSPHRASE`

### Engineering / pipeline (we own — can be wired dormant now)
- [ ] `checksums.txt` GPG signing (cheapest; do first) + publish public key
- [ ] Linux `.deb`/`.rpm` GPG signing (reuses the same key)
- [ ] macOS sign + notarize (`rcodesign` on Linux) — turns on when the Apple secrets land
- [ ] Windows Authenticode signing — turns on when the Windows signing creds land
- [ ] (Optional) macOS `.pkg` for offline stapling; real apt/yum repo

## Open decisions
1. macOS: **bare signed binary** (online Gatekeeper) vs. **`.pkg`** (offline). Recommend bare to start.
2. Windows: **OV vs EV vs Azure Trusted Signing**. Recommend Azure Trusted Signing.
3. Linux: **detached signatures + published key** vs. a **hosted apt/yum repo**. Recommend the former to start.
4. Where to run signing: **Linux runners** (`rcodesign` / `osslsigncode`) to keep CI cheap, vs. platform-native runners.

## References
- Apple Developer ID / notarization: <https://developer.apple.com/developer-id/>
- `rcodesign` (sign + notarize from Linux): <https://gregoryszorc.com/docs/apple-codesign/main/>
- Azure Trusted Signing: <https://learn.microsoft.com/azure/trusted-signing/>
- GoReleaser signing / notarize / nfpms signature docs: <https://goreleaser.com/customization/>
