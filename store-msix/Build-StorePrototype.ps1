# BookTriggerSaver - experimental Microsoft Store MSIX build.
# Separate from the GitHub Releases installer. Requires Go and Windows SDK (makeappx).
param([string]$OutputDirectory = "dist/store-msix")
$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest
Push-Location (Split-Path -Parent $PSScriptRoot)
try {
    $root = (Get-Location).Path
    $target = Join-Path $root $OutputDirectory
    $layout = Join-Path $target "layout"
    New-Item -ItemType Directory -Path (Join-Path $layout "Assets") -Force | Out-Null

    # Rebuild from source: do not modify or re-sign the public installer.
    & (Join-Path $root "Build.ps1") -Output (Join-Path $target "BookTriggerSaver.exe")
    if ($LASTEXITCODE -ne 0) { throw "Compilation failed" }

    # Build a separate Store-only executable. The public GitHub build remains untouched.
    & go build -tags store -buildvcs=false -trimpath -ldflags "-H=windowsgui -s -w" -o (Join-Path $target "BookTriggerSaver.exe") .
    if ($LASTEXITCODE -ne 0) { throw "Store-edition compilation failed" }
    # The classic builder already executes Windows GUI tests.
    # Compile/check the Store variant without rerunning flaky GDI+ GUI tests.
    & go test -tags store -run '^$' ./...
    if ($LASTEXITCODE -ne 0) { throw "Store-edition compilation checks failed" }
    & go vet -tags store ./...
    if ($LASTEXITCODE -ne 0) { throw "Store-edition static analysis failed" }
    $makeappx = Get-ChildItem "${env:ProgramFiles(x86)}\Windows Kits\10\bin" -Filter makeappx.exe -Recurse -ErrorAction SilentlyContinue |
        Where-Object { $_.FullName -match '\\x64\\makeappx\.exe$' } |
        Sort-Object FullName -Descending | Select-Object -First 1
    if (-not $makeappx) { throw "Windows SDK MakeAppx.exe was not found." }

    Copy-Item (Join-Path $target "BookTriggerSaver.exe") (Join-Path $layout "BookTriggerSaver.exe") -Force
    # Keep the screensaver inside the package; never copy the classic installer.
    Copy-Item (Join-Path $target "BookTriggerSaver.exe") (Join-Path $layout "BookTrigger.scr") -Force

    Add-Type -AssemblyName System.Drawing
    $source = [System.Drawing.Image]::FromFile((Join-Path $root "assets\logo.png"))
    try {
        foreach ($size in @(44, 150, 50)) {
            $bitmap = New-Object System.Drawing.Bitmap($size, $size)
            try {
                $g = [System.Drawing.Graphics]::FromImage($bitmap)
                try {
                    $g.Clear([System.Drawing.Color]::Transparent)
                    $g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
                    $g.DrawImage($source, 0, 0, $size, $size)
                } finally { $g.Dispose() }
                $name = switch ($size) { 44 {"Square44x44Logo.png"} 150 {"Square150x150Logo.png"} 50 {"StoreLogo.png"} }
                $bitmap.Save((Join-Path $layout "Assets\$name"), [System.Drawing.Imaging.ImageFormat]::Png)
            } finally { $bitmap.Dispose() }
        }
    } finally { $source.Dispose() }

    $manifest = @'
<?xml version="1.0" encoding="utf-8"?>
<Package xmlns="http://schemas.microsoft.com/appx/manifest/foundation/windows10"
 xmlns:uap="http://schemas.microsoft.com/appx/manifest/uap/windows10"
 xmlns:uap10="http://schemas.microsoft.com/appx/manifest/uap/windows10/10"
 xmlns:rescap="http://schemas.microsoft.com/appx/manifest/foundation/windows10/restrictedcapabilities"
 IgnorableNamespaces="uap uap10 rescap">
 <Identity Name="J.M.Contenidos.BookTriggerSaver"
           Publisher="CN=3E8CA7B6-6E05-45F0-8BD7-37C8E925922D"
           Version="11.5.0.0" ProcessorArchitecture="x64"/>
 <Properties>
  <DisplayName>BookTriggerSaver</DisplayName>
  <PublisherDisplayName>J.M. Contenidos</PublisherDisplayName>
  <Description>Salvapantallas literario de BookTrigger</Description>
  <Logo>Assets\StoreLogo.png</Logo>
 </Properties>
 <Resources><Resource Language="es-es"/></Resources>
 <Dependencies>
  <TargetDeviceFamily Name="Windows.Desktop" MinVersion="10.0.19041.0" MaxVersionTested="10.0.22000.0"/>
 </Dependencies>
 <Applications>
  <Application Id="App" Executable="BookTriggerSaver.exe"
    uap10:RuntimeBehavior="packagedClassicApp" uap10:TrustLevel="mediumIL">
   <uap:VisualElements DisplayName="BookTriggerSaver" Description="Salvapantallas literario"
     BackgroundColor="transparent"
     Square150x150Logo="Assets\Square150x150Logo.png"
     Square44x44Logo="Assets\Square44x44Logo.png"/>
  </Application>
 </Applications>
 <Capabilities><rescap:Capability Name="runFullTrust"/></Capabilities>
</Package>
'@
    [System.IO.File]::WriteAllText((Join-Path $layout "AppxManifest.xml"), $manifest, (New-Object System.Text.UTF8Encoding($false)))
    $package = Join-Path $target "BookTriggerSaver_11.5.0.0_x64_unsigned.msix"
    if (Test-Path $package) { Remove-Item $package -Force }
    & $makeappx.FullName pack /d $layout /p $package /o
    if ($LASTEXITCODE -ne 0) { throw "MakeAppx pack failed" }
    # Confirm both entry points survived packaging.
    $verify = Join-Path $target "verify"
    & $makeappx.FullName unpack /p $package /d $verify /o
    if ($LASTEXITCODE -ne 0) { throw "MSIX unpack verification failed" }
    foreach ($name in @("BookTriggerSaver.exe", "BookTrigger.scr", "AppxManifest.xml", "Assets\\StoreLogo.png")) {
        if (-not (Test-Path (Join-Path $verify $name))) { throw "Required MSIX entry missing: $name" }
    }
    Write-Host "Created and verified: $package"
    Write-Warning "PROTOTYPE ONLY. Uninstall/activation and Store validation are NOT yet verified."
} finally { Pop-Location }
