# BookTriggerSaver

**Salvapantallas literario gratuito para Windows (64 bits).** Convierte los momentos de inactividad del ordenador en una biblioteca visual: estanterías oscuras, portadas de libros, obras destacadas, un reloj en movimiento y detalles vegetales.

BookTriggerSaver muestra libros en español del catálogo de **BookTrigger**. Es una aplicación de **J.M. Contenidos**, disponible en **Microsoft Store**.

## Salvapantallas de libros en español

Disfruta de una **biblioteca virtual en tu ordenador** mientras no lo utilizas. BookTriggerSaver es un salvapantallas gratuito para Windows con portadas de libros en español, estanterías, obras destacadas y un reloj animado. El catálogo se actualiza automáticamente y guarda las imágenes ya consultadas para poder mostrarlas después sin conexión.

## Imágenes del salvapantallas

![BookTriggerSaver: presentación visual del salvapantallas literario](docs/images/img_1_1793248765696.jpg)

![BookTriggerSaver: biblioteca virtual con portadas de libros en español](docs/images/apps.45260.13510874358595962.12751690-5c74-4b1d-a159-8dcac98ea092.jpeg)

## Descargar

**Opción recomendada: [Descargar BookTriggerSaver desde Microsoft Store](https://apps.microsoft.com/detail/9NBLC0ZVZ08P).** Consulta la ficha oficial, las capturas de pantalla y realiza la instalación desde la tienda de Windows. Esta versión se distribuye mediante el proceso oficial de firma y publicación de Microsoft.

**[VER EN MICROSOFT STORE · DESCARGA OFICIAL](https://apps.microsoft.com/detail/9NBLC0ZVZ08P)**

**Versión alternativa de GitHub:** [descargar la versión beta v11.5 desde Releases](https://github.com/jmconteni2/BookTriggerSaver/releases/tag/v11.5). Consulta también las [instrucciones de descarga](DOWNLOAD.md).

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

## Apoya BookTriggerSaver

Si te gusta este salvapantallas literario, puedes apoyar el proyecto pulsando **Star** (la estrella) en la parte superior de este repositorio. Tu apoyo ayuda a que otras personas descubran BookTriggerSaver.

[![Estrellas del repositorio](https://img.shields.io/github/stars/jmconteni2/BookTriggerSaver?style=social)](https://github.com/jmconteni2/BookTriggerSaver/stargazers)

## Licencia y privacidad

Código y recursos propios bajo [licencia MIT](LICENSE), con aclaraciones en [NOTICE.md](NOTICE.md). Las portadas obtenidas del catálogo pertenecen a sus respectivos titulares y no forman parte del repositorio. Consulta la [política de privacidad](PRIVACY.md).
