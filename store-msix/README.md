# BookTriggerSaver — Microsoft Store (prototipo v11.5.0.3)

Rama aislada `store-msix-prototype`; la versión pública de GitHub, en `main`, no se modifica.

## Verificado en Windows por el usuario: versión 11.5.0.1

- Instalación MSIX con firma temporal de prueba.
- Activación automática de Windows al minuto y cierre por movimiento de ratón.
- Reaparición tras un nuevo minuto de inactividad.
- Portadas, interacción por Intro y enlaces correctos.
- Desactivar y restaurar desde la aplicación; posteriormente, desinstalación desde Configuración.
- Retirada de dos certificados temporales de prueba.

Desde 11.5.0.2 los recursos gráficos del paquete MSIX se generan desde el ICO original `winres/#1_0000.ico`, no desde el logo de J.M. Pocket. Aún necesita validación visual en el equipo del usuario.

## Nuevo mecanismo de recuperación (experimental 11.5.0.3)

MSIX no permite lanzar arbitrariamente un desinstalador propio al quitar la aplicación desde Windows. Si el archivo `.scr` reside en `LocalState`, al borrar los datos del paquete puede quedar una referencia `SCRNSAVE.EXE` rota.

Como prueba técnica, un ayudante Win32 sin empaquetar guarda una copia `BookTrigger.scr` y una copia de seguridad de los ajustes en `%LOCALAPPDATA%\BookTriggerSaverStore`. Se guarda también la identidad exacta de familia MSIX. En su siguiente ejecución, el `.scr` comprueba con `GetPackagesByPackageFamily` si la app continúa instalada. Si no lo está, restaura los valores previos, no muestra el salvapantallas y termina. La desactivación voluntaria restaura y elimina los archivos externos cuando no están abiertos.

**Este mecanismo no restaura la configuración inmediatamente en el momento de desinstalar**: lo hace en el siguiente arranque del `.scr`, o ejecutando `COMPROBAR-RETIRADA-DIRECTA.cmd`. Los archivos de salvaguarda podrían permanecer después de una desinstalación directa hasta su limpieza. El flujo es experimental y también hay que verificar si Microsoft acepta la copia del ejecutable fuera del paquete en la certificación.

## Prueba de desinstalación directa pendiente

1. Descargar el artefacto `BookTriggerSaver-MSIX-LOCAL-TEST-ONLY` del flujo más reciente. Descomprimir.
2. Ejecutar `INSTALAR-PRUEBA.cmd` y activar el salvapantallas como en la primera prueba.
3. Verificar que aparece automáticamente al minuto.
4. **Sin usar «Desactivar y restaurar»**, desinstalar BookTriggerSaver desde Configuración > Aplicaciones. Esto simula el problema real.
5. Ejecutar `COMPROBAR-RETIRADA-DIRECTA.cmd` desde el ZIP para forzar el siguiente arranque del `.scr` y verificar si Windows dejó de señalarlo.
6. Ejecutar `LIMPIAR-CERTIFICADOS-PRUEBA.cmd` para eliminar el certificado de esta nueva prueba.

No publicar hasta superar la prueba anterior, verificar actualizaciones MSIX, revisar identidad Name/Publisher en Partner Center y aprobar certificación. La firma MSIX de Microsoft se aplicará durante la distribución Store, no a una copia aislada del `.scr` ni al EXE clásico.
