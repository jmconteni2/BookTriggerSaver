# Create a disposable development-signed MSIX for local testing only.
# Never upload this signed package to Partner Center or distribute it publicly.
$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest
Push-Location (Split-Path -Parent $PSScriptRoot)
try {
    $output = Join-Path (Get-Location).Path "dist/store-msix"
    $source = Join-Path $output "BookTriggerSaver_11.5.1.0_x64_unsigned.msix"
    if (-not (Test-Path $source)) { throw "First build the unsigned Store MSIX." }
    $target = Join-Path $output "test"
    New-Item -ItemType Directory -Path $target -Force | Out-Null
    $msix = Join-Path $target "BookTriggerSaver_11.5.1.0_x64_TEST_ONLY.msix"
    Copy-Item $source $msix -Force

    $publisher = "CN=3E8CA7B6-6E05-45F0-8BD7-37C8E925922D"
    $certificate = New-SelfSignedCertificate -Type CodeSigningCert -Subject $publisher -CertStoreLocation "Cert:\CurrentUser\My" -NotAfter (Get-Date).AddDays(30)
    if (-not $certificate) { throw "Unable to generate the temporary test signing certificate." }
    try {
        Export-Certificate -Cert $certificate -FilePath (Join-Path $target "BookTriggerSaver-TEST.cer") -Force | Out-Null
        $signTool = Get-ChildItem "${env:ProgramFiles(x86)}\Windows Kits\10\bin" -Filter signtool.exe -Recurse -ErrorAction SilentlyContinue |
            Where-Object { $_.FullName -match '\\x64\\signtool\.exe$' } |
            Sort-Object FullName -Descending | Select-Object -First 1
        if (-not $signTool) { throw "Windows SDK SignTool was not found." }
        & $signTool.FullName sign /fd SHA256 /sha1 $certificate.Thumbprint $msix
        if ($LASTEXITCODE -ne 0) { throw "Development signing failed." }
    } finally {
        Remove-Item ("Cert:\CurrentUser\My\" + $certificate.Thumbprint) -ErrorAction SilentlyContinue
    }
    Copy-Item (Join-Path $PSScriptRoot "Install-Test.cmd") (Join-Path $target "INSTALAR-PRUEBA.cmd") -Force
    Copy-Item (Join-Path $PSScriptRoot "Remove-Test.cmd") (Join-Path $target "RETIRAR-PRUEBA.cmd") -Force
    Copy-Item (Join-Path $PSScriptRoot "Diagnose-Test.cmd") (Join-Path $target "COMPROBAR-ACTIVACION.cmd") -Force
    Copy-Item (Join-Path $PSScriptRoot "Remove-TestCertificate.cmd") (Join-Path $target "LIMPIAR-CERTIFICADOS-PRUEBA.cmd") -Force
    Copy-Item (Join-Path $PSScriptRoot "Check-DirectUninstall.cmd") (Join-Path $target "COMPROBAR-RETIRADA-DIRECTA.cmd") -Force
    Write-Host "Local-only signed test package and public test certificate created."
} finally { Pop-Location }
