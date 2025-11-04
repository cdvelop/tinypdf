package tinypdf

import (
	"github.com/cdvelop/tinypdf/fpdf"
)

type TinyPDF struct {
	Fpdf          *fpdf.Fpdf
	logger        func(message ...any)
	fontLoader    func(fontPath string) ([]byte, error)
	rootDirectory string
	fontsDirName  string
}

// Log escribe mensajes de log según el entorno
// En backend usa fmt.Println, en frontend usa console.log
func (tp *TinyPDF) Log(message ...any) {
	if tp.logger != nil {
		tp.logger(message...)
	}
}

func New(options ...any) *TinyPDF {

	tp := &TinyPDF{
		rootDirectory: ".",
		fontsDirName:  "fonts",
	}

	// Extraer rootDirectory y fontsDirName de las opciones
	for _, opt := range options {
		switch v := opt.(type) {
		case fpdf.RootDirectoryType:
			tp.rootDirectory = string(v)
		case fpdf.FontsDirName:
			tp.fontsDirName = string(v)
		}
	}

	// Inicializar las funciones de IO según el entorno
	tp.initIO()

	// Crear instancia de Fpdf con las opciones y las funciones de IO
	options = append(options, fpdf.WriteFileFunc(tp.writeFile))
	options = append(options, fpdf.ReadFileFunc(tp.readFile))
	options = append(options, fpdf.FileSizeFunc(tp.fileSize))
	options = append(options, tp.fontLoader)

	tp.Fpdf = fpdf.New(options...)

	return tp
}
