@echo off
setlocal
title BookTriggerSaver - Verificar desinstalacion directa
echo.
echo COMPROBACION DEL SALVAPANTALLAS TRAS DESINSTALAR DESDE WINDOWS
echo.
powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "$ErrorActionPreference='Stop'; $p=Get-AppxPackage -Name 'J.M.Contenidos.BookTriggerSaver' -ErrorAction SilentlyContinue; if($p){throw 'BookTriggerSaver sigue instalado. Desinstalalo desde Configuracion de Windows antes de esta comprobacion.'}; $d=Join-Path $env:LOCALAPPDATA 'BookTriggerSaverStore'; $scr=Join-Path $d 'BookTrigger.scr'; $before=(Get-ItemProperty -Path 'HKCU:\Control Panel\Desktop' -Name 'SCRNSAVE.EXE' -ErrorAction SilentlyContinue).'SCRNSAVE.EXE'; if($before -and $before.Equals($scr,[StringComparison]::OrdinalIgnoreCase)){ if(-not (Test-Path -LiteralPath $scr -PathType Leaf)){throw 'ERROR: Windows apunta a un SCR que ya no existe'}; Write-Host 'Activando comprobacion automatica de retirada...'; & $scr /s; Start-Sleep -Milliseconds 500 }; $after=(Get-ItemProperty -Path 'HKCU:\Control Panel\Desktop' -Name 'SCRNSAVE.EXE' -ErrorAction SilentlyContinue).'SCRNSAVE.EXE'; if($after -and $after.Equals($scr,[StringComparison]::OrdinalIgnoreCase)){throw 'ERROR: No se ha restaurado la configuracion anterior'}; if($after -and -not (Test-Path -LiteralPath $after -PathType Leaf)){throw ('ERROR: El salvapantallas actual no existe: '+$after)}; Write-Host ('Salvapantallas tras retirar MSIX: '+$after); Write-Host 'CORRECTO: no quedan referencias a BookTriggerSaver en el salvapantallas activo' -ForegroundColor Green"
if errorlevel 1 (
 echo.
 echo LA COMPROBACION HA DETECTADO UN PROBLEMA. No borres archivos.
)
echo.
pause
endlocal
