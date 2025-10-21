package repositories

// PDFRepository define la interfaz para operaciones con archivos PDF
type PDFRepository interface {
	// ExtractText extrae el texto de un archivo PDF
	ExtractText(pdfPath string) (string, error)
}
