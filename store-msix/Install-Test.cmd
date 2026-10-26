@echo off
setlocal
title BookTriggerSaver - Instalacion de prueba
cd /d "%~dp0"
fltmc >nul 2>&1
if errorlevel 1 (
 echo Esta prueba necesita permiso de administrador para confiar en el certificado temporal.
 powershell.exe -NoProfile -Command "Start-Process -FilePath '%~f0' -Verb RunAs"
 exit /b
)
echo.
echo PRUEBA LOCAL DE BOOKTRIGGERSAVER PARA MICROSOFT STORE
echo.
echo Se instalara una version DE PRUEBA y su certificado temporal.
echo No cambiara tu salvapantallas hasta que abras la aplicacion y lo actives.
echo No es la version destinada a publicar.
echo.
choice /M "Quieres instalar esta version de prueba"
if errorlevel 2 exit /b 0
powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "$ErrorActionPreference='Stop'; Import-Certificate -FilePath '.\BookTriggerSaver-TEST.cer' -CertStoreLocation 'Cert:\LocalMachine\TrustedPeople' | Out-Null; Add-AppxPackage -Path '.\BookTriggerSaver_11.5.0.1_x64_TEST_ONLY.msix' -ErrorAction Stop; Write-Host 'INSTALACION DE PRUEBA COMPLETADA' -ForegroundColor Green"
if errorlevel 1 (
 echo.
 echo La instalacion ha fallado. Haz una captura de la ventana.
) else (
 echo.
 echo Abre BookTriggerSaver desde el menu Inicio para probarlo.
 echo Para desinstalarlo, primero usa Desactivar y restaurar dentro de la aplicacion.
)
echo.
pause
endlocal
