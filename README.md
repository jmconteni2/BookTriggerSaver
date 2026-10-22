# BookTriggerSaver

Salvapantallas gratuito y de código abierto para Windows x64 de J.M. Pocket. Muestra el catálogo de BookTrigger sobre estanterías oscuras, con portadas destacadas, reloj móvil y decoración vegetal.

## Descargar e instalar

Consulta [Descarga v11.5](DOWNLOAD.md). El instalador actual está **sin firma Authenticode**; la solicitud a SignPath está en preparación, todavía no enviada ni aprobada.

Ejecuta `BookTriggerSaver_v11_5.exe`, elige el tiempo de inactividad y pulsa **Instalar y activar**. El botón **Ejecutar ahora** permite probarlo al terminar. La instalación configura el salvapantallas del usuario actual sin permisos de administrador.

- Activación configurable: 1, 2, 3, 5, 10, 15 o 30 minutos.
- Libro destacado durante 15 segundos, incluidos fundidos suaves de medio segundo.
- Caja del reloj y logotipos en movimiento cada 30 segundos.
- Intro abre el enlace del libro visible en el navegador predeterminado; las demás teclas cierran el salvapantallas.
- Catálogo consultado cada dos minutos, con imágenes descargadas en segundo plano y caché local para trabajar sin conexión después de la primera sincronización.

Los archivos se instalan en `%APPDATA%\JMPocketSalvapantallas`. El instalador cambia la selección, activación y tiempo de espera del salvapantallas; guarda la configuración anterior. No cambia la contraseña ni la protección de bloqueo de Windows. Para quitarlo, usa **Aplicaciones instaladas > BookTriggerSaver > Desinstalar**. La desinstalación restaura los ajustes anteriores cuando corresponde y conserva la caché.

## Compilación y verificación

Con Go 1.27.3 en Windows, ejecuta `./Build.ps1`. El script ejecuta pruebas nativas y `go vet`, genera recursos con `go-winres` 0.3.3 y compila `dist/BookTriggerSaver_v11_5.exe`. La aplicación instalada no necesita Go ni dependencias adicionales. Las únicas bibliotecas de ejecución son la biblioteca estándar de Go y las de Windows.

[GitHub Actions](https://github.com/jmconteni2/BookTriggerSaver/actions) ejecuta pruebas portables con detector de carreras y pruebas nativas de Windows. La [compilación candidata v11.5](https://github.com/jmconteni2/BookTriggerSaver/actions/runs/39824074037) superó los 25 grupos de pruebas sin omisiones, incluidas las pruebas del Explorador, argumentos Unicode, enlace de la portada visible y dibujo GDI/GDI+. El responsable del proyecto ha comunicado que funciona correctamente en su última prueba y ha aprobado el diseño. Esto no sustituye pruebas en todos los equipos Windows.

La apertura se delega una sola vez al Explorador de Windows existente, después de cerrar la ventana del salvapantallas. No usa el auxiliar ni la espera de escritorio de versiones anteriores. El registro local `apertura.log` incluye fases, dominio y hash, sin la URL completa ni sus parámetros.

## Code signing policy

Consulta la [política de firma y estado de la solicitud](SIGNPATH.md). Actualmente no existe firma ni patrocinio aprobado por SignPath Foundation.

## Licencia y privacidad

Código y recursos propios distribuidos bajo [MIT](LICENSE), con las aclaraciones de [componentes y recursos](NOTICE.md). Las portadas obtenidas del catálogo pertenecen a sus respectivos titulares y no se incluyen en este repositorio. Consulta la [política de privacidad](PRIVACY.md).
