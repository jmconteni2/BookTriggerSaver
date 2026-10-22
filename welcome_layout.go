package main

type welcomeLayout struct {
	Width, Height, Margin, PosterY, PosterWidth, PosterHeight, FooterY, Scale int
}

func completionLayout(workW, workH, dpi int) welcomeLayout {
	if dpi <= 0 {
		dpi = 96
	}
	scale := func(v int) int { return (v*dpi + 48) / 96 }
	margin, header, footer := scale(20), scale(88), scale(86)
	pw := scale(751)
	if max := workW - 2*margin - scale(48); pw > max {
		pw = max
	}
	ph := pw * 475 / 751
	if max := workH - scale(64) - header - footer; ph > max {
		ph = max
		pw = ph * 751 / 475
	}
	return welcomeLayout{pw + 2*margin, header + ph + footer, margin, header, pw, ph, header + ph, dpi}
}

const welcomeText = `¡Bienvenido al salvapantallas de BookTrigger!

Mientras tu pantalla descansa, las historias siguen vivas.

Has abierto las puertas de una biblioteca muy especial, un rincón donde los libros, la imaginación y la tranquilidad se dan la mano.

Nuestra biblioteca cobrará vida ante tus ojos. Los libros irán cambiando y, de vez en cuando, una portada se iluminará para invitarte a descubrir una nueva historia.

Porque nunca se sabe dónde puede estar esperando tu próxima lectura favorita.

¡Disfruta de tu pequeño rincón literario!

¿Cómo funciona?

• INTRO: Descubre el libro destacado en tu navegador.
• Ratón o cualquier otra tecla: Regresa a tu escritorio.

BookTrigger. Un proyecto de J.M. Contenidos con la colaboración de J.M. Pocket.`
