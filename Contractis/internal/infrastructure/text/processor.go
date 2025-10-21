package text

import (
	"strings"

	"github.com/rodascaar/contractis/internal/domain/entities"
)

// Processor implementa el procesamiento de texto
type Processor struct{}

// NewProcessor crea una nueva instancia de Processor
func NewProcessor() *Processor {
	return &Processor{}
}

// SplitText divide el texto en fragmentos manejables
func (p *Processor) SplitText(text string, maxSize int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{""}
	}

	// Validar tamaño mínimo
	if maxSize < entities.MinChunkSize {
		maxSize = entities.MinChunkSize
	}

	// Si el texto es menor que maxSize, retornar como un solo chunk
	if len(text) <= maxSize {
		return []string{text}
	}

	var chunks []string
	for len(text) > 0 {
		// Determinar el tamaño del chunk actual
		chunkSize := maxSize
		if len(text) < maxSize {
			chunkSize = len(text)
		}

		// Buscar punto de división natural
		splitPoint := chunkSize
		if chunkSize == maxSize && len(text) > maxSize {
			searchStart := (chunkSize * 7) / 10 // Buscar desde 70% del chunk
			if idx := strings.LastIndexAny(text[searchStart:chunkSize], "\n\n"); idx != -1 {
				splitPoint = searchStart + idx + 2
			} else if idx := strings.LastIndexAny(text[searchStart:chunkSize], ".!?"); idx != -1 {
				splitPoint = searchStart + idx + 1
			} else if idx := strings.LastIndexAny(text[searchStart:chunkSize], " \n\t"); idx != -1 {
				splitPoint = searchStart + idx + 1
			}
		}

		// Extraer chunk y limpiar
		chunk := strings.TrimSpace(text[:splitPoint])
		if len(chunk) > 0 {
			chunks = append(chunks, chunk)
		}

		// Avanzar en el texto
		text = strings.TrimSpace(text[splitPoint:])
	}

	return chunks
}

// CleanFragment elimina emojis, símbolos innecesarios y formato redundante
func (p *Processor) CleanFragment(text string) string {
	// Eliminar marcadores de encabezado
	text = strings.ReplaceAll(text, "###", "")
	text = strings.ReplaceAll(text, "##", "")
	text = strings.ReplaceAll(text, "#", "")

	// Eliminar emojis comunes
	emojis := []string{"📄", "✅", "❌", "⚠️", "🔍", "📊", "💡", "🎯", "🚀", "⚡", "🔧"}
	for _, emoji := range emojis {
		text = strings.ReplaceAll(text, emoji, "")
	}

	// Eliminar líneas de separación
	text = strings.ReplaceAll(text, "---", "")
	text = strings.ReplaceAll(text, "***", "")
	text = strings.ReplaceAll(text, "===", "")

	// Eliminar espacios múltiples
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	// Eliminar saltos de línea múltiples
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}

	return strings.TrimSpace(text)
}

// EstimateTokens estima la cantidad de tokens
func (p *Processor) EstimateTokens(text string, charsPerToken int) int {
	if charsPerToken <= 0 {
		charsPerToken = entities.CharsPerToken
	}
	return len(text) / charsPerToken
}
