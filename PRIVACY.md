# Privacidad y conexiones

BookTriggerSaver consulta el endpoint público de BookTrigger para recibir títulos, direcciones de portadas y enlaces de compra. Descarga las imágenes desde las direcciones que devuelve ese catálogo. Estos servicios y servidores reciben los datos normales de una petición HTTP, como la dirección IP y el identificador `BookTriggerSaver/11.0`.

El programa no incorpora analítica, publicidad añadida, registro de pulsaciones, identificadores de seguimiento propios, acceso a documentos del usuario ni envío de capturas. El teclado y el ratón solo se usan mientras está abierta la ventana del salvapantallas para salir o abrir el enlace elegido. Al pulsar Intro, Windows abre el enlace del libro en el navegador y se aplican las condiciones del sitio de destino.

El catálogo y las imágenes se conservan en `%APPDATA%\JMPocketSalvapantallas`. La caché nueva usa `catalog-v11.json` y el subdirectorio `covers`. Los errores de instalación o ejecución pueden quedar en `instalacion.log`, dentro del mismo directorio. No se envían esos registros.

El instalador guarda la configuración anterior del salvapantallas y configura su selección, activación y tiempo de inactividad para el usuario actual. La desinstalación restaura los ajustes anteriores cuando corresponde y conserva la caché. No se modifica la protección por contraseña de Windows.

Las conexiones forman parte de la función anunciada de sincronizar la biblioteca; no se escribe en Google Sheets. El backend de Google Apps Script y los servidores de imágenes son servicios externos. El proyecto no afirma controlar sus políticas de privacidad ni el tratamiento que realizan esos proveedores.

La revisión 11.2 registra localmente las fases de apertura de compra en `apertura.log`, con versión, hora, identificadores de proceso, dominio e identificador hash del enlace, sin URL completa ni parámetros. Este registro no se envía y se reinicia al superar 256 KiB.

Desde v11.5 no se inicia ningún auxiliar para abrir la compra ni se añade ningún proceso residente. Se delega la apertura al Explorador de Windows existente. Las fases se registran localmente como `explorer_*`, con el mismo dominio e identificador hash y sin enviar el registro.
