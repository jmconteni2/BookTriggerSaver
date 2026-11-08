@echo off
setlocal
chcp 65001 >nul
cd /d "%~dp0"
title Capturas reales de BookTriggerSaver
echo.
echo CAPTURAS AUTOMATICAS DEL SALVAPANTALLAS
echo.
echo No toques el raton ni el teclado mientras se muestran los libros.
echo Se guardaran cuatro capturas reales y se cerrara el salvapantallas.
echo.
start "" /wait "%~dp0BookTriggerCapturas.exe" --capturar
if errorlevel 1 (
    echo.
    echo No se han completado las capturas. Revisa ERROR.txt en la carpeta de imagenes.
) else (
    echo.
    echo Capturas completadas.
)
start "" "%USERPROFILE%\Pictures\Capturas BookTriggerSaver"
pause
