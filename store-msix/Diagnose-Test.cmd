@echo off
setlocal
title BookTriggerSaver - Diagnostico de prueba
echo.
echo DIAGNOSTICO DEL SALVAPANTALLAS EN WINDOWS
echo.
reg query "HKCU\Control Panel\Desktop" /v "SCRNSAVE.EXE"
reg query "HKCU\Control Panel\Desktop" /v "ScreenSaveActive"
reg query "HKCU\Control Panel\Desktop" /v "ScreenSaveTimeOut"
echo.
powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "$p=Get-AppxPackage -Name 'J.M.Contenidos.BookTriggerSaver'; if(-not $p){Write-Host 'MSIX no encontrado'; exit}; Write-Host ('MSIX instalado: '+$p.Version); $selected=(Get-ItemProperty -Path 'HKCU:\Control Panel\Desktop' -Name 'SCRNSAVE.EXE' -ErrorAction SilentlyContinue).'SCRNSAVE.EXE'; Write-Host ('Archivo seleccionado existe: '+[bool](Test-Path -LiteralPath $selected -PathType Leaf)); $d=Join-Path $env:LOCALAPPDATA ('Packages\'+$p.PackageFamilyName+'\LocalState'); Write-Host ('SCR estable existe: '+[bool](Test-Path -LiteralPath (Join-Path $d 'BookTrigger.scr') -PathType Leaf)); $log=Join-Path $d 'inicio_salvapantallas.log'; if(Test-Path $log){Write-Host 'Ultimas ejecuciones reales:';Get-Content $log -Tail 8}else{Write-Host 'Sin registros de arranque del archivo SCR'}"
echo.
echo Copia las ultimas lineas si sigue sin activarse.
pause
