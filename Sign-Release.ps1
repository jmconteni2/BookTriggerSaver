param(
    [Parameter(Mandatory=$true)][string]$CertificateThumbprint,
    [Parameter(Mandatory=$true)][string]$TimestampUrl,
    [string]$SignTool = "signtool.exe",
    [string]$Output = "dist/BookTriggerSaver_v11_firmado.exe"
)
$ErrorActionPreference = "Stop"
Push-Location $PSScriptRoot
try {
    & "$PSScriptRoot/Build.ps1" -Output $Output
    # The installer and the installed .scr are the same signed file.
    & $SignTool sign /sha1 $CertificateThumbprint /fd SHA256 /tr $TimestampUrl /td SHA256 $Output
    if ($LASTEXITCODE -ne 0) { throw "No se pudo firmar el ejecutable." }
    & $SignTool verify /pa /v $Output
    if ($LASTEXITCODE -ne 0) { throw "La firma del ejecutable no es válida." }
    Get-FileHash -Algorithm SHA256 $Output | Format-List | Out-File -Encoding utf8 "$Output.sha256.txt"
} finally { Pop-Location }
