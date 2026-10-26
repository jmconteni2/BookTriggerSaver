@echo off
setlocal
title BookTriggerSaver - Prueba MSIX
cd /d "%~dp0"
echo.
echo PRUEBA DE BOOKTRIGGERSAVER PARA MICROSOFT STORE
echo.
echo Solo instala la aplicacion de prueba.
echo No cambia el salvapantallas activo hasta que abras BookTriggerSaver.
echo Se anadira un certificado temporal de prueba a tu cuenta Windows.
echo.
choice /M "Quieres continuar con la prueba"
if errorlevel 2 exit /b 0
powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "$ErrorActionPreference='Stop'; Import-Certificate -FilePath '.\BookTriggerSaver-TEST.cer' -CertStoreLocation 'Cert:\CurrentUser\TrustedPeople' | Out-Null; Add-AppxPackage -Path '.\BookTriggerSaver_11.5.0.0_x64_TEST_ONLY.msix' -ErrorAction Stop; Write-Host 'INSTALACION DE PRUEBA COMPLETADA' -ForegroundColor Green"
if errorlevel 1 (
 echo.
 echo La instalacion ha fallado. Haz una captura de este mensaje.
) else (
 echo.
 echo Abre BookTriggerSaver desde el menu Inicio para probarlo.
 echo Antes de desinstalarlo, abre la aplicacion y usa Desactivar y restaurar.
)
echo.
pause
endlocal
