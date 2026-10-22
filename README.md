# BookTriggerSaver v11 — motor reconstruido desde fuentes

Salvapantallas Windows x64 de J.M. Pocket. El motor, el instalador y el fundido se compilan íntegramente desde los archivos Go de este proyecto. El resultado es un único ejecutable que también se instala como `BookTrigger.scr`.

No se utiliza `original_v9.exe`, `engine.scr`, un parche binario, el adaptador `btfade1.dll` ni un compilador de C. No se han recuperado las fuentes originales: esta es una implementación nueva del comportamiento y del diseño medidos en v10.3.

## Instalar

Ejecuta `BookTriggerSaver_v11_2.exe`, elige los minutos de inactividad y pulsa **Instalar y activar**. La pantalla final mantiene el cartel original y el botón **Ejecutar ahora**. La actualización se puede instalar sobre v10.3 con el salvapantallas cerrado.

- Tiempo de espera: 1, 2, 3, 5, 10, 15 o 30 minutos.
- Cada libro destacado permanece seleccionado durante 15 segundos, incluidos los fundidos de entrada y salida de medio segundo.
- La caja del logo y la hora cambia de posición cada 30 segundos.
- Intro abre el enlace de la portada destacada. Las demás teclas y la actividad del ratón cierran el salvapantallas.
- La vista previa de Windows se aloja en su ventana padre y no consume el teclado del panel de configuración.
- Las descargas se realizan en segundo plano. Un fallo de conexión conserva la biblioteca guardada y no interrumpe la selección vigente.
- La instalación es por usuario y no necesita permisos de administrador. No altera la contraseña ni las condiciones de desbloqueo de Windows.

La primera ejecución de v11 necesita conexión para crear su nueva caché. No se borran las cachés ni los archivos de las versiones anteriores. El origen del catálogo se conserva en `endpoint.txt` dentro de `%APPDATA%\JMPocketSalvapantallas`; el conector de Google Apps Script sigue siendo el mismo y no se modifica ni despliega desde este proyecto.

## Compilar en Windows

Instala Go 1.27.3 para reproducir la versión entregada y ejecuta:

```powershell
./Build.ps1
```

El script ejecuta las pruebas, genera los recursos desde `winres/winres.json` y el icono, y compila `dist/BookTriggerSaver_v11_2.exe`. La primera generación de recursos descarga la herramienta abierta `go-winres` 0.3.3. No requiere comprar herramientas ni certificados. La aplicación terminada no necesita Go instalado.

`go.mod` fija Go 1.24 como versión mínima del lenguaje. La versión del compilador de cada compilación queda registrada junto al ejecutable.

## Pruebas

```bash
go test -race -count=1 ./...
```

Las pruebas portables comparan la geometría con mediciones guardadas de v10.3, los tiempos, el recorrido de la caja, las asociaciones portada/enlace, la vista previa y el manejo del catálogo y la caché. No utilizan el ejecutable antiguo para ejecutarse. Las referencias son datos JSON en `testdata`.

En Windows se ejecuta además `TestNativeGraphicsOffscreen`: usa GDI/GDI+ reales para dibujar en memoria, comprobar el interior de la caja y la asociación de la portada dibujada con su enlace. No instala el salvapantallas ni abre el navegador. Esta prueba está preparada y compilada, pero no se ha ejecutado en el entorno Linux de esta entrega.

`.github/workflows/build.yml` contiene las pruebas portables y la compilación y pruebas nativas en Windows. El flujo queda listo para ejecutarse cuando se publique el repositorio; no se ha ejecutado en GitHub ni se ha creado un repositorio desde esta entrega.

## Fuentes

- `scene.go`: caja, mosaico, selección de libros y curva del fundido.
- `catalog.go`: cliente HTTP, descarga de portadas y caché.
- `graphics_windows.go`: dibujo nativo, imágenes, tipografía, recursos y doble búfer.
- `saver_windows.go`: ventana completa o vista previa, temporizador y entradas del usuario.
- `main_windows.go`: instalación, configuración, desinstalación y lanzamiento.
- `welcome_windows.go`: cartel final original.
- `winres`: definición editable de recursos y metadatos de Windows.

## Firma

La entrega está **sin firma Authenticode**. Reconstruir las fuentes no crea un certificado ni elimina los avisos de Windows.

La preparación para solicitar la firma gratuita se explica en `SIGNPATH.md`. SignPath debe revisar y aceptar el proyecto. El archivo firmado del instalador se copia sin modificar como `.scr`, de modo que no hace falta firmar un motor externo ni una DLL propia.

`Sign-Release.ps1` permite compilar, firmar y verificar el único ejecutable si se dispone de un certificado compatible con SignTool. No crea certificados ni cambia la confianza del equipo. La integración concreta con SignPath se configura después de que exista y se apruebe la cuenta del proyecto.

## Licencias y privacidad

El código nuevo tiene licencia MIT. Los recursos y las dependencias se describen en `NOTICE.md`. Consulta `PRIVACY.md` para conocer las conexiones y los datos que utiliza la aplicación.

## Corrección v11.1

La carga inicial espera un primer grupo de hasta ocho portadas antes de mostrar resultados parciales. El cambio de destacado conserva la distribución del mosaico y las actualizaciones del catálogo conservan la identidad de cada portada aunque cambie el orden de descarga. Esta revisión necesita validación visual en el equipo donde se detectó el fallo.

La apertura de compra usa ShellExecuteExW con inicialización COM y espera de activación antes de terminar. Intro conserva el enlace de la última portada dibujada correctamente, ignora la repetición de una tecla Intro mantenida al arrancar y muestra el error de Windows si no puede abrir el navegador.

## Apertura de compra v11.2

Intro inicia una instancia auxiliar del mismo ejecutable en `winsta0\default`. Esta espera a que termine el proceso del salvapantallas y a que el escritorio normal esté activo antes de abrir el enlace. No cambia la contraseña, no cambia de escritorio y no desbloquea Windows. Si no se desbloquea en dos minutos, la petición caduca. La petición al navegador tiene un límite de 30 segundos y no se reintenta automáticamente.

`%APPDATA%\JMPocketSalvapantallas\apertura.log` registra versión, hora, proceso, fases, dominio y un identificador hash del enlace; no guarda la URL completa ni sus parámetros. La aceptación de Windows no acredita que la página web haya terminado de cargar.
