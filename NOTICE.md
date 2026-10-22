# Componentes y recursos

El código de esta reconstrucción se distribuye bajo la licencia MIT incluida en `LICENSE`.

La aplicación usa la biblioteca estándar y el runtime de Go. Su licencia BSD se incluye en `GO-LICENSE.txt`. No hay módulos de terceros en `go.mod`. La herramienta de compilación `go-winres`, versión 0.3.3, genera los recursos de Windows desde el JSON y el icono; no forma parte del código que se ejecuta en el salvapantallas.

Windows aporta GDI+, GDI, User32, Kernel32, Shell32, Advapi32 y Comctl32. Estas bibliotecas del sistema no se distribuyen con el proyecto.

El logotipo, el halo, el cartel de bienvenida y el icono son los recursos facilitados para el proyecto de J.M. Pocket. Se conservan sus bytes. La licencia de código no concede derechos sobre marcas comerciales ni sobre contenidos de terceros. Antes de publicar estos recursos debe confirmarse que se dispone de sus derechos de distribución bajo las condiciones del proyecto abierto.

Las portadas, títulos y enlaces se descargan del catálogo público configurado y no forman parte del repositorio. Sus derechos pertenecen a sus respectivos titulares. El proyecto no incluye libros, manuscritos ni la hoja de cálculo de origen.

`assets/ivy.png` es un adorno vegetal generado para este proyecto con la herramienta integrada de imágenes. El recurso se distribuye junto al proyecto bajo MIT, sin atribuir derechos sobre marcas ajenas. La estantería se dibuja desde el código y no incorpora la fotografía de referencia ni sus libros.
