# BookTriggerSaver — Microsoft Store (experimental branch)

This branch is deliberately isolated from the published v11.5 GitHub release and its signed installer.

## Progress

- Reusable Windows SDK packaging script: `store-msix/Build-StorePrototype.ps1`.
- Exact Microsoft Store package identity and publisher.
- Executable application entry point (the previous capture-based package had none).
- Classic desktop/full-trust MSIX declaration, with bundled logos.
- New build from Go source rather than copying user-specific AppData, logs, cached covers or browser files.
- Package remains **unsigned** for eventual Partner Center processing.

## Important: not ready for publication

The existing Go executable runs its own installer, writes `%APPDATA%\JMPocketSalvapantallas`, changes `HKCU\Control Panel\Desktop` screensaver settings and registers its own uninstaller. A Store-installed MSIX has a different lifecycle. In particular, removing the MSIX may leave a selected `.scr` path to a file that has disappeared, and cached files may be redirected by package virtualization. These behaviors need actual testing and source changes before release.

Before publishing:
1. Test the package on Windows, including first launch, activation, preview, screen timeout and opening books.
2. Verify catalog / covers downloading and cache across relaunch and version updates.
3. Implement safe uninstall/reset of the Windows screensaver setting for the MSIX lifecycle, without disturbing the classic GitHub installer.
4. Test upgrading and removing the MSIX on an isolated Windows account or VM.
5. Validate with Windows App Certification Kit and Partner Center ingestion.

Do **not** merge this branch or upload the prototype as a production Store package before those checks.

## Build (Windows)

Requires working Go toolchain, repository source, and Windows SDK with MakeAppx. Run PowerShell from the repository root:

```powershell
.\store-msix\Build-StorePrototype.ps1
```

Expected prototype output: `dist/store-msix/BookTriggerSaver_11.5.0.0_x64_unsigned.msix`.

The production installer already published on GitHub is not modified by this script.
