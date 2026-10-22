# Code signing policy

## Estado

Solicitud al programa gratuito de SignPath Foundation **en preparación, no enviada ni aprobada**. BookTriggerSaver v11.5 se distribuye sin firma Authenticode. No existe actualmente patrocinio, certificado configurado ni integración de firma activada.

## Responsabilidades

El responsable del proyecto y propietario del repositorio es [jmconteni2](https://github.com/jmconteni2), de J.M. Pocket. Mantiene el código y los scripts, revisa las contribuciones externas y decidirá personalmente qué versiones solicita firmar. Actualmente no hay otros mantenedores designados. No se atribuye una revisión independiente a una segunda persona.

El responsable ha indicado que activó la autenticación de dos factores en GitHub. Si se aprueba la cuenta, deberá activar también MFA en SignPath y aprobar manualmente cada solicitud de firma. No se solicitarán firmas automáticamente desde contribuciones externas.

## Origen de los binarios

El código, recursos editables y scripts de compilación son públicos y tienen licencia MIT, con las licencias y aclaraciones indicadas en [NOTICE.md](NOTICE.md). No se incluyen motores externos opacos. GitHub Actions compila desde el checkout del repositorio, ejecuta pruebas portables y nativas y publica el ejecutable y su SHA-256.

La [candidata v11.5](DOWNLOAD.md) corresponde al commit `2fa5aa5fd4fe6adcd52ed7c5f1ce1cc8131ea064`, con [compilación y pruebas verificadas](https://github.com/jmconteni2/BookTriggerSaver/actions/runs/39824074037). El instalador y el salvapantallas son el mismo ejecutable: se copia sin modificación como `BookTrigger.scr`.

## Política prevista tras la aceptación

Se configurará la integración de origen verificable con GitHub Actions, restricciones de nombre y versión del producto y aprobación manual de cada versión por el responsable. Solo se publicará como firmado un archivo cuya firma se haya verificado realmente.

El crédito requerido, **a publicar como crédito activo únicamente después de la aprobación**, es: “Free code signing provided by [SignPath.io](https://about.signpath.io), certificate by [SignPath Foundation](https://signpath.org)”. Ese certificado identificaría a SignPath Foundation como editor; no sería un certificado a nombre legal de J.M. Pocket.

## Privacidad

Consulta [PRIVACY.md](PRIVACY.md) para las conexiones de catálogo, imágenes y enlaces, los datos locales y los servicios de terceros. La instalación anuncia y configura la selección y el tiempo de espera del salvapantallas; la desinstalación está disponible desde Windows. No se cambia la protección por contraseña.

## Condiciones oficiales

- [Condiciones de SignPath Foundation](https://signpath.org/terms.html)
- [Formulario de solicitud](https://signpath.org/apply)

SignPath decide la aceptación, incluidas las comprobaciones de reputación y control del proyecto. No hay aprobación garantizada ni fecha de activación confirmada. Una firma válida tampoco garantiza que Windows deje de mostrar todos los avisos de reputación.
