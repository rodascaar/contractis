package services

// TextProcessor define la interfaz para procesamiento de texto
type TextProcessor interface {
	// SplitText divide el texto en fragmentos manejables
	SplitText(text string, maxSize int) []string

	// CleanFragment limpia un fragmento de texto
	CleanFragment(text string) string

	// EstimateTokens estima la cantidad de tokens
	EstimateTokens(text string, charsPerToken int) int
}
