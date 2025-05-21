package env

var (
	// logger is a function that will be used for logging
	// depending on the environment (WASM or backend)
	Logger = SetupDefaultLogger()
	// FileWriter is a function that will be used for writing PDF data to a file
	// eg: FileWriter("output.pdf", data)
	FileWriter = SetupDefaultFileWriter()
)
