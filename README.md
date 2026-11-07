# BookTriggerSaver

**Salvapantallas literario gratuito para Windows (64 bits).** Convierte los momentos de inactividad del ordenador en una biblioteca visual: estanterías oscuras, portadas de libros, obras destacadas, un reloj en movimiento y detalles vegetales.

BookTriggerSaver muestra libros del catálogo de **BookTrigger**. Es una aplicación de **J.M. Contenidos**, distribuida también a través de **Microsoft Store**.

### Literary bookshelf screensaver for Windows

**BookTriggerSaver** is a free **literary screensaver for Windows 64-bit** featuring a virtual bookshelf, book cover artwork, rotating featured books and a moving clock. It brings a digital library to your screen when the computer is idle. The online book catalog updates automatically and previously loaded images remain available from a local cache.

**Looking for a book screensaver, bookshelf screensaver or book-cover screensaver for Windows?** See the features below and find the installation options in the download section.

## Descargar

**Opción recomendada: [Descargar BookTriggerSaver desde Microsoft Store](https://apps.microsoft.com/detail/9NBLC0ZVZ08P).** Consulta la ficha oficial, las capturas de pantalla y realiza la instalación desde la tienda de Windows. Esta versión se distribuye mediante el proceso oficial de firma y publicación de Microsoft.

**[VER EN MICROSOFT STORE · DESCARGA OFICIAL](https://apps.microsoft.com/detail/9NBLC0ZVZ08P)**

**Versión de GitHub:** [instrucciones de descarga](DOWNLOAD.md). Esta compilación independiente se distribuye sin firma Authenticode y Windows puede mostrar una advertencia de seguridad.

## ¿Qué muestra el salvapantallas?

- Estanterías con portadas de libros del catálogo de BookTrigger.
- Un libro destacado que cambia cada 15 segundos, con transiciones suaves.
- Un reloj y elementos gráficos que cambian de posición periódicamente.
- Acceso al libro visible mediante la tecla Intro, que abre su enlace en el navegador.
- Actualización del catálogo cada dos minutos y caché local de imágenes para su posterior uso sin conexión.

## Instalación y uso de la versión de GitHub

Ejecuta `BookTriggerSaver_v11_5.exe`, selecciona el tiempo de inactividad y pulsa **Instalar y activar**. Puedes probarlo con **Ejecutar ahora**.

La activación puede configurarse en 1, 2, 3, 5, 10, 15 o 30 minutos. El programa se instala para el usuario actual, sin requerir permisos de administrador.

Los archivos se guardan en `%APPDATA%\JMPocketSalvapantallas`. La instalación conserva una copia de la configuración anterior del salvapantallas y no modifica la contraseña ni la protección de bloqueo de Windows. Para desinstalarlo, usa **Aplicaciones instaladas > BookTriggerSaver > Desinstalar**.

## Código fuente y verificación

Este repositorio contiene el código fuente. Para compilarlo en Windows con Go 1.27.3, ejecuta `./Build.ps1`. El script realiza pruebas, ejecuta `go vet` y genera `dist/BookTriggerSaver_v11_5.exe`. El programa instalado no requiere Go.

[GitHub Actions](https://github.com/jmconteni2/BookTriggerSaver/actions) ejecuta pruebas automatizadas. La [compilación candidata v11.5](https://github.com/jmconteni2/BookTriggerSaver/actions/runs/39824074037) superó las pruebas documentadas; estos resultados no sustituyen pruebas en todos los equipos.

## Licencia y privacidad

Código y recursos propios bajo [licencia MIT](LICENSE), con aclaraciones en [NOTICE.md](NOTICE.md). Las portadas obtenidas del catálogo pertenecen a sus respectivos titulares y no forman parte del repositorio. Consulta la [política de privacidad](PRIVACY.md).
