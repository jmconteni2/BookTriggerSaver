# Preparación para SignPath Foundation

## Qué está resuelto

El instalador y el motor se generan desde código fuente de este paquete. No existe una dependencia de ejecutables anteriores ni de una DLL opaca. El resultado es un único PE de Windows x64, con recursos generables desde sus definiciones, metadatos de producto y versión, pruebas portables y una prueba gráfica preparada para Windows.

Se incluye licencia MIT del código y documentación de privacidad, componentes, compilación e instalación. El flujo de GitHub Actions compila desde un checkout del repositorio y entrega un artefacto de compilación sin firmar. No dispone de credenciales ni ejecuta ninguna solicitud de firma.

## Lo que falta para solicitarlo

1. Confirmar la distribución abierta del código y los derechos de los recursos gráficos antes de hacer público el repositorio.
2. Crear el repositorio público de J.M. Pocket y subir las fuentes. El ZIP de esta entrega no equivale a haberlo publicado.
3. Ejecutar el flujo de Windows, superar su prueba gráfica nativa y revisar el resultado en Windows 10/11: vista previa, dibujo, activación por inactividad y apertura real del navegador.
4. Definir las personas que mantienen, revisan y aprueban versiones, activar la autenticación multifactor y establecer las reglas de revisión del repositorio.
5. Solicitar la aceptación del proyecto en https://signpath.org/apply y facilitar los enlaces públicos del proyecto y de sus versiones.
6. Tras la aprobación, configurar la integración de SignPath con los artefactos de GitHub, la política de firma y la aprobación manual de cada versión.

SignPath exige una licencia abierta admitida, ausencia de componentes propietarios no permitidos y trazabilidad verificable entre las fuentes y los binarios. También valora la reputación del proyecto. Disponer de fuentes completas no implica que la solicitud esté aceptada ni garantiza que lo vaya a estar.

La firma gratuita de la fundación utiliza un certificado emitido a SignPath Foundation. No equivale a obtener un certificado gratuito a nombre legal de J.M. Pocket. No se debe anunciar que SignPath firma el proyecto hasta que la solicitud esté aprobada y la firma se haya aplicado realmente.

Cuando se apruebe el servicio, la página pública deberá incorporar la política de firma y los créditos que exija SignPath. La presente entrega no publica esos créditos como si ya existiera patrocinio.

## Documentación oficial consultada

- Condiciones: https://signpath.org/terms.html
- Servicio: https://signpath.org/
- Firma y reputación en Windows: https://learn.microsoft.com/en-us/windows/apps/package-and-deploy/smartscreen-reputation
