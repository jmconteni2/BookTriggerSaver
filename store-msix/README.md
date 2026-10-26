# BookTriggerSaver — edición Microsoft Store (rama experimental)

La rama `store-msix-prototype` está separada de `main`. **No modifica el instalador público v11.5.**

## Estado real comprobado (26 de octubre de 2026)

- El usuario instaló el MSIX de prueba 11.5.0.1 con certificado temporal.
- Activación automática de Windows al minuto: funciona en su equipo.
- Movimiento del ratón para cerrar, reactivación automática y apertura de fichas con Intro: comprobados.
- «Desactivar y restaurar» y desinstalación desde Configuración de Windows: comprobados.
- Los iconos MSIX se generaban por error desde `assets/logo.png` (J.M. Pocket), mientras que el EXE original utiliza `winres/#1_0000.ico`. Corregido desde 11.5.0.2.
- No se ha comprobado todavía en la máquina del usuario que el nuevo icono aparezca en Inicio; la validación automatizada solo verifica el empaquetado.

## Arquitectura

El MSIX lleva `BookTriggerSaver.exe` y `BookTrigger.scr`. La interfaz instalada empaquetada almacena una copia estable del `.scr` en `LocalState` del paquete e invoca un ejecutable auxiliar sin virtualización del registro para configurar realmente el salvapantallas de Windows. El tiempo, selección y estado anteriores se conservan en `configuracion_anterior.json`.

La compilación MSIX ejecuta las pruebas específicas de Store, `go vet` y verifica los ejecutables y los recursos incluidos. Los iconos del menú Inicio y Store se generan desde el archivo ICO original del programa.

## Artefactos del flujo de GitHub

- **BookTriggerSaver-MSIX-LOCAL-TEST-ONLY**: ZIP firmado con certificado autofirmado temporal. Solo para pruebas locales, incluye `INSTALAR-PRUEBA.cmd`, `COMPROBAR-ACTIVACION.cmd`, `RETIRAR-PRUEBA.cmd` y `LIMPIAR-CERTIFICADOS-PRUEBA.cmd`.
- **BookTriggerSaver-MSIX-EXPERIMENTAL-NOT-FOR-PUBLICATION**: contiene el MSIX sin firmar, candidato a enviar por Partner Center solo una vez resueltas las limitaciones pendientes.

## Pendientes antes de Microsoft Store

1. **Desinstalación directa**: MSIX no ejecuta desinstaladores personalizados. El usuario puede desinstalar desde Configuración sin abrir «Desactivar y restaurar». Esto podría dejar `SCRNSAVE.EXE` apuntando a un archivo eliminado. Debe solucionarse o someterse a evaluación expresa antes de publicar.
2. **Actualización y limpieza**: verificar que los archivos `.scr` en `LocalState` se actualizan de forma segura y que no permanecen restos tras retirar el paquete.
3. Verificar con Partner Center que la identidad `Name` y el `Publisher` del manifiesto corresponden exactamente a la reserva real del producto, y justificar `runFullTrust` en certificación.
4. Validar la publicación y las políticas de Microsoft Store. La compilación y la prueba local no garantizan la aprobación.

**El paquete publicado en Microsoft Store recibirá la firma de Microsoft. No requiere comprar certificado propio para distribuirlo dentro de la Store.** No subir la versión firmada para pruebas ni divulgar el certificado temporal.
