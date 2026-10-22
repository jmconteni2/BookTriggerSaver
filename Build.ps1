param([string]$Output = "dist/BookTriggerSaver_v11_4.exe")
$ErrorActionPreference = "Stop"
Push-Location $PSScriptRoot
try {
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    go test -count=1 ./...
    if ($LASTEXITCODE -ne 0) { throw "Las pruebas de Windows han fallado." }
    go vet ./...
    if ($LASTEXITCODE -ne 0) { throw "El análisis de Go ha fallado." }
    go run github.com/tc-hib/go-winres@v0.3.3 make --in winres/winres.json --arch amd64 --out rsrc
    if ($LASTEXITCODE -ne 0) { throw "No se pudieron generar los recursos desde sus fuentes." }
    $OutputParent = Split-Path -Parent $Output
    if ($OutputParent) { New-Item -ItemType Directory -Force $OutputParent | Out-Null }
    go build -buildvcs=false -trimpath -ldflags "-H=windowsgui -s -w" -o $Output .
    if ($LASTEXITCODE -ne 0) { throw "La compilación ha fallado." }
    go version | Set-Content -Encoding utf8 "$Output.toolchain.txt"
    Get-FileHash -Algorithm SHA256 $Output | Format-List | Out-File -Encoding utf8 "$Output.sha256.txt"
} finally { Pop-Location }
