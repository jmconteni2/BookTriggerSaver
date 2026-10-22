# Componentes y recursos

El código de esta reconstrucción se distribuye bajo la licencia MIT incluida en `LICENSE`.

La aplicación usa la biblioteca estándar y el runtime de Go. Su licencia BSD se incluye en `GO-LICENSE.txt`. No hay módulos de terceros en `go.mod`. La herramienta de compilación `go-winres`, versión 0.3.3, genera los recursos de Windows desde el JSON y el icono; no forma parte del código que se ejecuta en el salvapantallas.

Windows aporta GDI+, GDI, User32, Kernel32, Shell32, Advapi32 y Comctl32. Estas bibliotecas del sistema no se distribuyen con el proyecto.

Los logotipos de J.M. Pocket y BookTrigger, el halo, el cartel de bienvenida y el icono fueron facilitados por el responsable del proyecto, quien autorizó su publicación junto al código bajo MIT. Los archivos suministrados se conservan sin modificar. Esta licencia no concede derechos sobre marcas comerciales ni sobre contenidos de terceros.

Las portadas, títulos y enlaces se descargan del catálogo público configurado y no forman parte del repositorio. Sus derechos pertenecen a sus respectivos titulares. El proyecto no incluye libros, manuscritos ni la hoja de cálculo de origen.

`assets/ivy.png` es un adorno vegetal generado para este proyecto con la herramienta integrada de imágenes. El recurso se distribuye junto al proyecto bajo MIT, sin atribuir derechos sobre marcas ajenas. La estantería se dibuja desde el código y no incorpora la fotografía de referencia ni sus libros.
