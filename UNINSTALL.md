# Uninstall changes, version 0.1.0-0004

Upgrade to this version BEFORE uninstalling. Build through GitHub Actions as before.
The new English/Czech uninstall wizard offers an unchecked permanent deletion option.
Selecting it runs explicit cleanup of package-owned /volumeN/@appdata/PrusaConnectLocal
contents only. Cleanup is skipped during upgrades and without explicit confirmation.
Unexpected paths or mounted data are rejected. The DSM-managed directory itself is not removed.
External backups, printer files, browser credentials, browser certificate exceptions,
DSM system logs and any custom external storage/database are NOT deleted.
Not selecting the option only disables this script's cleanup; DSM data preservation
is not guaranteed. Back up before uninstalling. Deletion cannot be undone.

No actual deletion on NAS has been tested. Wizard rendering and lifecycle permissions
must be verified on DSM. No compiled SPK is included.
