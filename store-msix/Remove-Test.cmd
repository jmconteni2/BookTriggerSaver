@echo off
setlocal
title BookTriggerSaver - Retirar prueba
cd /d "%~dp0"
fltmc >nul 2>&1
if errorlevel 1 (
 powershell.exe -NoProfile -Command "Start-Process -FilePath '%~f0' -Verb RunAs"
 exit /b
)
echo.
echo RETIRAR LA VERSION DE PRUEBA DE BOOKTRIGGERSAVER
echo.
echo Si activaste el salvapantallas, primero abre BookTriggerSaver
echo y pulsa Desactivar y restaurar, para recuperar el anterior.
echo.
choice /M "Quieres retirar ahora la version de prueba"
if errorlevel 2 exit /b 0
powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "$ErrorActionPreference='Stop'; $p=Get-AppxPackage -Name 'J.M.Contenidos.BookTriggerSaver' -ErrorAction SilentlyContinue; if($p) { $v=(Get-ItemProperty -Path 'HKCU:\Control Panel\Desktop' -Name 'SCRNSAVE.EXE' -ErrorAction SilentlyContinue).'SCRNSAVE.EXE'; $stable=Join-Path $env:LOCALAPPDATA ('Packages\'+$p.PackageFamilyName+'\LocalState\BookTrigger.scr'); if($v -and ($v.Equals($stable,[StringComparison]::OrdinalIgnoreCase) -or $v.StartsWith($p.InstallLocation,[StringComparison]::OrdinalIgnoreCase))) { throw 'Aun esta activo este salvapantallas. Abre BookTriggerSaver y elige Desactivar y restaurar antes de quitarlo.' }; Remove-AppxPackage -Package $p.PackageFullName -ErrorAction Stop }; $c=Get-PfxCertificate -FilePath '.\BookTriggerSaver-TEST.cer'; Remove-Item ('Cert:\LocalMachine\TrustedPeople\'+$c.Thumbprint) -ErrorAction SilentlyContinue; Write-Host 'PRUEBA RETIRADA' -ForegroundColor Green"
if errorlevel 1 (
 echo La retirada no se ha completado. Haz una captura de esta ventana.
)
echo.
pause
endlocal
