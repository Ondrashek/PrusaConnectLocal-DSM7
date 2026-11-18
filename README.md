# Prusa Connect Local for Synology DSM 7

**Experimental source prototype, not a hardware-tested release.** Target: DS220+
(geminilake, Linux amd64). No Docker, SSH or Task Scheduler is required on the NAS.
This archive contains source code and a GitHub Actions build, NOT a compiled SPK.

## What is implemented

- Native upstream server, pinned to `v0.1.0-beta.1`.
- GitHub Actions downloads `prusa_connect_amd64` and verifies official SHA256SUMS.
- Package Center start/stop, child supervision and shutdown of child extensions.
- HTTPS browser access on **18443**, with a generated self-signed certificate.
- Protected setup-code portal at `https://<NAS-address>:18443/_synology/setup`.
  This URL is included in the Package Center description. The username is `admin`;
  the password is chosen in the package installation wizard, NOT your DSM password.
- English/Czech wrapper portal and package description; upstream UI translations
  remain whatever the official upstream release provides.
- Launcher uses the browser's actual DSM hostname and port, not 5000/5001.
  The wrapper also reads `admin_port` and `secure_admin_port` from
  `/etc/synoinfo.conf` for diagnostic display. It does not modify DSM settings.
- Backend port **18080** is HTTP for printers; browser access is proxied over HTTPS.
- Dedicated unprivileged package account, private state, no upstream secret logs.

## Build and install using only graphical interfaces

1. Create a GitHub repository and upload the CONTENTS of this archive, including
   `.github/workflows/build.yml` (hidden directories must be included).
   If GitHub web upload omits `.github`, create the workflow file using the web editor.
2. Open **Actions → Build experimental DSM7 package → Run workflow**.
3. Download the artifact `PrusaConnectLocal-DSM7-experimental` and extract the SPK.
4. DSM: **Package Center → Manual Install**, select the SPK. Review the warning
   for this unsigned third-party experimental package. Choose a unique random
   portal password, at least 8 characters. Do not reuse any existing password.
5. Open the package. The launcher shows HTTPS application and setup-code links.
   The first HTTPS visit warns about a self-signed certificate. Verify you are
   connecting to your own NAS before accepting it.
6. Open the setup-code portal, log in as `admin` with the installation password,
   copy the one-time code, then open the application and complete upstream setup.
   Choose authenticated access and create the Prusa administrator account.
7. Add a printer in Prusa Connect. Verify its exported connection settings use
   the NAS LAN address, port **18080**, and `tls=false` for this prototype.

No NAS shell commands or scheduled tasks are involved. If DSM firewall is enabled,
allow the chosen ports from the trusted LAN in the DSM GUI. The package deliberately
does not weaken your firewall or open router ports. QuickConnect/path-based external
reverse proxies are not supported by the launcher. Use the NAS's direct LAN address.

## Important limitations

- This code has NOT been installed or tested on a DS220+ or any DSM system.
  GitHub Actions compilation, upstream download, wizard compatibility, permissions,
  desktop integration and real printer operation still need end-to-end validation.
- The supplied source is not a promise that an SPK will install unchanged. Treat the
  first build as a development candidate, not a production release.
- HTTP 18080 exposes the full upstream application/API as well as printer routes.
  It is not a printer-only isolated listener. Do not expose either port to the internet.
  HTTPS browser support does not imply encrypted printer traffic.
- The generated certificate is not automatically trusted. It contains the NAS names
  and addresses available at first start. If the address changes or the certificate
  expires, replace `https.crt` and `https.key` in package state with a matching PEM
  pair while stopped, then restart. Certificate upload/renewal UI is not implemented.
- No automatic free-port selection: 18443/18080 must be available. DSM's own port
  is discovered automatically; these application ports are intentionally fixed.
- The setup portal uses its own Basic authentication over HTTPS, not DSM SSO.
  A unique high-entropy password is required. No MFA or persistent lockout is provided.
- The setup code is kept only in RAM and displayed for at most 30 minutes after
  capture. It can remain visible after upstream setup consumes it, but is no longer
  valid. Restart via Package Center to obtain a new code if setup is still pending.
- The wrapper has no complete log viewer or backup/restore GUI. Back up package state
  before updates. Uninstall may delete state; preservation on uninstall is not promised.
- Upstream self-update is disabled. Update via a rebuilt SPK after backing up data.
- No package capability is granted for UDP 123. Upstream's optional time-server
  feature may report permission denied; this integration does not replace DSM NTP.
- Upstream is public beta and can change configuration and APIs.

## Development

`go test -race ./...` tests parsing and helper logic.
`CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/pcl-wrapper ./cmd/pcl-wrapper`

Place the pinned upstream LICENSE beside the verified upstream executable.
Build script accepts an existing wrapper and verified upstream executable:
`python3 tools/build_spk.py --wrapper build/pcl-wrapper --upstream build/upstream/prusa_connect_amd64 --output build/PrusaConnectLocal.spk`

`python3 tools/test_package.py build/PrusaConnectLocal.spk` checks archive structure,
JSON and checksum, not DSM behavior. Runtime has no external Go dependencies.

## Acceptance checklist before release

- Build succeeds, Go race tests pass, ELF architecture and static linkage verified.
- Install on exact DSM version, wizard password transfer, package account ownership.
- Start/stop/reboot and crash restart; no orphan extensions after Stop.
- DSM ports changed from defaults; launcher still opens correct NAS.
- HTTPS UI, file upload, SSE, printer registration and WebSocket traffic.
- Setup portal denies unauthenticated access and does not persist the setup code.
- Printer connects over 18080; UI export settings manually verified.
- Upgrade retains configuration, accounts, files, portal password and TLS identity.
- Port conflicts fail visibly; firewall behavior verified; uninstall data policy agreed.
- Review upstream license and include all required notices before distributing SPKs.

## References

- Product: https://www.prusa3d.com/cs/p/connect/connect-local/
- Upstream: https://github.com/prusa3d/prusa-connect
- Configuration: https://github.com/prusa3d/prusa-connect/blob/main/docs/src/content/docs/config.md
- Release installation: https://github.com/prusa3d/prusa-connect/blob/main/docs/src/content/docs/install-from-a-release.md
- DSM package guide: https://help.synology.com/developer-guide/
- DSM privilege: https://help.synology.com/developer-guide/privilege/privilege_config.html
- DSM desktop config: https://help.synology.com/developer-guide/integrate_dsm/config.html
