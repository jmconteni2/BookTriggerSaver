# BookTriggerSaver — edición Microsoft Store (prototipo separado)

Esta rama NO sustituye la publicación v11.5 de GitHub ni modifica su instalador autofirmado.

## Qué prepara esta rama

- Compilación independiente con la etiqueta de Go `store`, sin el instalador clásico, su registro de desinstalación ni la copia tradicional de `.scr` en AppData.
- MSIX con `BookTriggerSaver.exe` (interfaz) y `BookTrigger.scr` (motor para Windows). Ambos se generan desde las fuentes de esta rama.
- Datos y caché separados en `BookTriggerSaverStore`.
- Antes de activarse se guarda el estado anterior del salvapantallas. La aplicación ofrece **Desactivar y restaurar**, que recupera la configuración si aún está seleccionada.
- El flujo automatizado comprueba la compilación de la edición Store, empaqueta el MSIX sin firmar y verifica que contiene ambos ejecutables y el manifiesto.

## Limitaciones que impiden publicar todavía

**MUY IMPORTANTE: el paquete es experimental, no está listo para Microsoft Store.**

1. Windows MSIX no ejecuta una acción de desinstalación personalizada. Si el usuario quita la aplicación directamente sin desactivarla primero, podría quedar en el registro un `SCRNSAVE.EXE` que apunta a un archivo eliminado. **La desactivación manual actual no es una solución definitiva.**
2. Las actualizaciones MSIX pueden instalar el paquete en una ruta versionada nueva. La selección anterior del `.scr` podría apuntar a la versión retirada. Necesita resolverse antes de publicar.
3. Falta verificar en una instalación Windows real si el sistema puede arrancar y previsualizar el `.scr` desde el directorio del paquete, incluido acceso a recursos, caché, enlaces y prevención de múltiples instancias.
4. Hay que validar la identidad del paquete con Partner Center, las capacidades restringidas y la certificación de Windows. La firma gratuita de Microsoft solo llega **después de aprobar la publicación**, no convierte un prototipo en instalador listo para utilizar.
5. El archivo MSIX que entrega GitHub permanece sin firma. **No instalarlo directamente** como si fuera un paquete de producción.

## Compilar y revisar

Al modificar esta rama, GitHub Actions ejecuta `.github/workflows/store-msix.yml`. El artefacto de pruebas se llama `BookTriggerSaver-MSIX-EXPERIMENTAL-NOT-FOR-PUBLICATION`.

Para compilar en un entorno Windows con Go y Windows SDK:

```powershell
.\store-msix\Build-StorePrototype.ps1
```

La versión pública firmada de GitHub se mantiene intacta.
