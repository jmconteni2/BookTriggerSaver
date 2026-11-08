@echo off
setlocal
title BookTriggerSaver - Limpiar certificado temporal
fltmc >nul 2>&1
if errorlevel 1 (
 echo Se requiere el permiso de administrador para retirar el certificado temporal.
 powershell.exe -NoProfile -Command "Start-Process -FilePath '%~f0' -Verb RunAs"
 exit /b
)
echo.
echo LIMPIAR CERTIFICADOS TEMPORALES DE BOOKTRIGGERSAVER
echo.
echo Este proceso solo retira certificados autofirmados de prueba
echo con la identidad exacta del paquete BookTriggerSaver.
echo No afecta a certificados de Microsoft ni de otros editores.
echo.
choice /M "Quieres retirar los certificados de prueba"
if errorlevel 2 exit /b 0
powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "$ErrorActionPreference='Stop'; $n='CN=3E8CA7B6-6E05-45F0-8BD7-37C8E925922D'; $certs=@(Get-ChildItem 'Cert:\LocalMachine\TrustedPeople' | Where-Object { $_.Subject -eq $n -and $_.Issuer -eq $n -and ($_.NotAfter - $_.NotBefore).TotalDays -le 32 }); foreach($c in $certs){ Remove-Item -LiteralPath ('Cert:\LocalMachine\TrustedPeople\'+$c.Thumbprint) -ErrorAction Stop }; Write-Host ('CERTIFICADOS DE PRUEBA RETIRADOS: '+$certs.Count) -ForegroundColor Green"
if errorlevel 1 (
 echo No se ha podido completar la limpieza. No borres certificados manualmente.
)
echo.
pause
endlocal
