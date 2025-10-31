package tinypdf

import (
	"github.com/cdvelop/tinypdf/fpdf"
)

type TinyPDF struct {
	Fpdf   *fpdf.Fpdf
	logger func(message ...any)
}

// Log escribe mensajes de log según el entorno
// En backend usa fmt.Println, en frontend usa console.log
func (tp *TinyPDF) Log(message ...any) {
	if tp.logger != nil {
		tp.logger(message...)
	}
}

func New(options ...any) *TinyPDF {

	tp := &TinyPDF{}

	// Inicializar las funciones de IO según el entorno
	tp.initIO()

	// Crear instancia de Fpdf con las opciones y las funciones de IO
	options = append(options, fpdf.WriteFileFunc(tp.writeFile))
	options = append(options, fpdf.ReadFileFunc(tp.readFile))
	options = append(options, fpdf.FileSizeFunc(tp.fileSize))

	tp.Fpdf = fpdf.New(options...)

	return tp
}
