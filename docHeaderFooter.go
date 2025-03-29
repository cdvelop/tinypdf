package tinypdf

import "strconv"

// AddHeader - add a header function, if present this will be automatically called by AddPage()
func (gp *GoPdf) AddHeader(f func()) {
	gp.headerFunc = f
}

// AddFooter - add a footer function, if present this will be automatically called by AddPage()
func (gp *GoPdf) AddFooter(f func()) {
	gp.footerFunc = f
}

// AddPageHeader agrega un encabezado al documento
func (d *Document) AddPageHeader(text string) *TextBuilder {
	// Usar el TextBuilder existente con estilo Normal
	builder := d.newTextBuilder(text, d.fontConfig.PageHeader, FontRegular)

	d.AddHeader(func() {
		d.SetY(15) // Posición Y para el encabezado
		builder.Draw()
	})

	// Devolver el TextBuilder para permitir encadenamiento de métodos
	return builder
}

// AddPageFooter agrega un pie de página al documento
func (d *Document) AddPageFooter(text string) *TextBuilder {
	// Usar el TextBuilder existente con estilo Footer
	builder := d.newTextBuilder("hola 1 "+text, d.fontConfig.PageFooter, FontRegular)
	builder.positioning = fixedPosition
	d.AddFooter(func() {
		// Posicionar en la parte inferior de la página
		pageHeight := d.config.PageSize.H
		// d.SetY(pageHeight - d.fontConfig.PageFooter.Size) // 20 puntos desde el borde inferior
		d.SetY(pageHeight) // 20 puntos desde el borde inferior
		builder.Draw()
	})

	// Devolver el TextBuilder para permitir encadenamiento de métodos
	return builder
}

//  add page number to the text builder
func (tb *TextBuilder) WithPageNumber() *TextBuilder {
	// Obtener el texto actual
	currentText := tb.text

	// Modificar el texto para incluir el número de página
	if currentText != "" {
		currentText += " "
	}
	currentText += strconv.Itoa(tb.doc.curr.IndexOfPageObj)

	// Actualizar el texto en el TextBuilder
	tb.text = currentText

	return tb
}
