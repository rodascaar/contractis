package pdf

import (
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

// Extractor implementa la extracción de texto de archivos PDF
type Extractor struct{}

// NewExtractor crea una nueva instancia de Extractor
func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractText extrae el texto de un archivo PDF
func (e *Extractor) ExtractText(pdfPath string) (string, error) {
	file, r, err := pdf.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("error al abrir el PDF: %w", err)
	}
	defer file.Close()

	var text strings.Builder
	totalPages := r.NumPage()

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		content, err := page.GetPlainText(nil)
		if err != nil {
			return "", fmt.Errorf("error al extraer texto de la página %d: %w", pageNum, err)
		}

		text.WriteString(content)
		text.WriteString(fmt.Sprintf("\n\n--- FIN DE PÁGINA %d ---\n\n", pageNum))
	}

	return text.String(), nil
}
