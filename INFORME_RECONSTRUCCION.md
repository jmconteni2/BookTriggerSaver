# Reconstrucción de BookTriggerSaver v11

## Resultado

Se ha escrito un motor nuevo de Go y se ha integrado en el mismo ejecutable del instalador. Su compilación requiere las fuentes, los recursos gráficos suministrados y herramientas abiertas. No utiliza el EXE original, una extracción o parche de instrucciones, ni el adaptador de fundido de v10.3.

El instalador conserva el selector de minutos, la activación y selección automáticas, la copia de configuración anterior y la pantalla final con el cartel original. La instalación genera `BookTrigger.scr` copiando su propio ejecutable. Por tanto, una firma aplicada al ejecutable final también se conserva en ese archivo instalado.

## Referencia visual

Se midieron la caja, sus cuatro capas redondeadas, el logo, el separador, el reloj y el texto inferior en las cinco posiciones, usando cuatro tamaños de pantalla. El código nuevo coincide con esas 20 referencias. El mosaico se contrastó con 26 referencias de coordenadas, que incluyen 1920×1080, 3840×2160, 1366×768, 800×600, 320×240 y una vista previa de 200×120.

Las referencias JSON se incluyen para poder repetir las comparaciones sin el motor antiguo. No son capturas de la interfaz nueva. Se conservan sin cambios los PNG del logo, del halo y del cartel, y el icono.

## Comportamiento

- Selección de 15 segundos con tiempo monotónico de Go. Los fundidos de 500 ms están incluidos en ese intervalo.
- Movimiento de la caja cada dos intervalos, conservando juntos el logo, el reloj y sus coordenadas relativas.
- Selección entre libros con un enlace HTTP/HTTPS válido, evitando repetir inmediatamente el anterior cuando existe otra opción.
- La asociación de título, portada y enlace se mantiene en un único registro. El enlace solicitado por Intro procede de la portada que la ventana ha presentado.
- Los cambios de catálogo se aplican al terminar el intervalo vigente; no interrumpen al libro destacado.
- Las descargas y la lectura de caché se realizan en segundo plano, sin llamadas gráficas fuera del hilo de Windows.
- Se limita el tamaño del catálogo, de cada descarga de imagen y de las dimensiones decodificadas. La caché gráfica tiene un presupuesto de 64 MiB; el fondo se reutiliza entre cambios para evitar redibujar todas las portadas durante cada cuadro del fundido.
- Doble búfer de GDI, dibujo de imágenes y transparencias mediante GDI+, y tipografía Segoe UI con las medidas anteriores.
- Vista previa como ventana hija, cierre por entrada del usuario y apertura de enlaces mediante el navegador de Windows.

## Validación y límites

Las 14 pruebas portables comprueban la geometría, la temporización, el fundido, el enlace del libro seleccionado, las actualizaciones durante una selección, los códigos de teclado, la caché, respuestas defectuosas, límites de tamaño y cancelación de descargas. Se ejecutaron además con el detector de carreras de Go.

Se ha compilado el ejecutable Windows x64 y se ha ejecutado `go vet` para ese destino. La prueba gráfica nativa de Windows está escrita y se compila, pero este entorno Linux no permite ejecutar una sesión real de Windows. Por ello, el aspecto final, el rendimiento en el equipo, la activación por inactividad, la vista previa nativa y la apertura efectiva del navegador quedan pendientes de esa comprobación. No se presentan las pruebas portables como integración completa con Windows.

La caché de v11 usa un formato propio y necesita una primera descarga. La actualización no borra archivos ni cachés de versiones anteriores. La desinstalación conserva el lanzador que la está ejecutando y la caché, igual que en la versión anterior.

## Estado de la firma

No se ha firmado el ejecutable, publicado un repositorio ni enviado una solicitud a SignPath. La reconstrucción elimina la dependencia del motor sin fuentes; la publicación, las condiciones de licencia y recursos, la aceptación y la firma siguen siendo pasos separados, descritos en `SIGNPATH.md`.

## Reconstrucción limpia comprobada

Se copió el proyecto a una carpeta temporal sin archivos EXE, SCR, DLL ni recursos SYso compilados. Se regeneraron los recursos a partir del JSON y del icono y se compiló con Go 1.27.3. El ejecutable obtenido coincide byte a byte con el entregado. Los detalles y el SHA-256 están en `BUILD_VERIFICATION.json`. `TESTS.log` recoge las 14 pruebas portables ejecutadas con el detector de carreras.
