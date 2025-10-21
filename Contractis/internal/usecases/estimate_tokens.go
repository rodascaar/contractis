package usecases

import (
	"fmt"
	"log"

	"github.com/rodascaar/contractis/internal/domain/entities"
	"github.com/rodascaar/contractis/internal/domain/repositories"
	"github.com/rodascaar/contractis/internal/domain/services"
)

// EstimateTokensUseCase maneja la estimación de tokens
type EstimateTokensUseCase struct {
	pdfRepo       repositories.PDFRepository
	textProcessor services.TextProcessor
}

// NewEstimateTokensUseCase crea una nueva instancia del caso de uso
func NewEstimateTokensUseCase(
	pdfRepo repositories.PDFRepository,
	textProcessor services.TextProcessor,
) *EstimateTokensUseCase {
	return &EstimateTokensUseCase{
		pdfRepo:       pdfRepo,
		textProcessor: textProcessor,
	}
}

// Execute ejecuta la estimación de tokens
func (uc *EstimateTokensUseCase) Execute(pdfPath string, maxTokensConfig int) (*entities.TokenEstimation, error) {
	// Extraer texto del PDF
	content, err := uc.pdfRepo.ExtractText(pdfPath)
	if err != nil {
		estimation := entities.NewTokenEstimation(0)
		estimation.SetError(err)
		return estimation, err
	}

	if content == "" {
		estimation := entities.NewTokenEstimation(0)
		estimation.SetError(fmt.Errorf("no se pudo extraer texto del PDF"))
		return estimation, fmt.Errorf("no se pudo extraer texto del PDF")
	}

	return uc.estimateTokens(content, maxTokensConfig), nil
}

func (uc *EstimateTokensUseCase) estimateTokens(text string, maxTokensConfig int) *entities.TokenEstimation {
	charCount := len(text)
	estimation := entities.NewTokenEstimation(charCount)

	// Estimación conservadora: ~3 caracteres por token
	estimatedInputTokens := charCount / entities.CharsPerToken

	// Para modelos online con mucho contexto, estimar procesamiento en una sola petición
	if estimatedInputTokens < (entities.OnlineContextWindow - entities.SafetyMargin - entities.MaxOutputTokens) {
		log.Printf("Estimando procesamiento en una sola petición para documento pequeño")
		return uc.estimateSingleRequest(text, maxTokensConfig)
	}

	// Procesamiento por chunks para documentos grandes
	log.Printf("Estimando procesamiento por chunks para documento grande")

	// Calcular chunks
	maxChunkSize := entities.DefaultChunkSize
	maxChunkTokens := entities.MaxInputTokens / 2
	if maxChunkSize/entities.CharsPerToken > maxChunkTokens {
		maxChunkSize = maxChunkTokens * entities.CharsPerToken
	}

	chunks := uc.textProcessor.SplitText(text, maxChunkSize)
	numChunks := len(chunks)

	// System prompt tokens
	systemPrompt := `Analiza contratos legales en español. Identifica: terminación unilateral, penalizaciones, jurisdicción, riesgos. Respuesta completa en español, sin emojis ni formato markdown.`
	systemPromptTokens := len(systemPrompt) / entities.CharsPerToken

	// FASE 1: Tokens por chunk
	phase1TokensPerChunk := systemPromptTokens + (maxChunkSize / entities.CharsPerToken) + entities.Phase1MaxTokens
	phase1TotalTokens := phase1TokensPerChunk * numChunks

	// FASE 2: Consolidación
	phase1OutputTotal := entities.Phase1MaxTokens * numChunks
	phase2InputTokens := systemPromptTokens + phase1OutputTotal

	phase2OutputTokens := maxTokensConfig
	if phase2OutputTokens < 800 {
		phase2OutputTokens = 800
	}
	if phase2OutputTokens > 2000 {
		phase2OutputTokens = 2000
	}

	phase2TotalTokens := phase2InputTokens + phase2OutputTokens
	totalTokens := phase1TotalTokens + phase2TotalTokens

	// Recomendación de maxTokens
	recommendedMaxTokens := uc.calculateRecommendedMaxTokens(numChunks, charCount)

	// Establecer valores en la estimación
	estimation.EstimatedTokens = estimatedInputTokens
	estimation.Chunks = numChunks
	estimation.SystemPromptTokens = systemPromptTokens
	estimation.Phase1Tokens = phase1TotalTokens
	estimation.Phase2InputTokens = phase2InputTokens
	estimation.Phase2OutputTokens = phase2OutputTokens
	estimation.TotalTokens = totalTokens
	estimation.RecommendedMaxTokens = recommendedMaxTokens

	// Advertencias
	warning := uc.generateWarning(numChunks, totalTokens)
	if warning != "" {
		estimation.SetWarning(warning)
	}

	return estimation
}

func (uc *EstimateTokensUseCase) estimateSingleRequest(text string, maxTokensConfig int) *entities.TokenEstimation {
	charCount := len(text)
	estimation := entities.NewTokenEstimation(charCount)

	// Estimación conservadora: ~3 caracteres por token
	estimatedInputTokens := charCount / entities.CharsPerToken

	// System prompt tokens
	systemPrompt := `Analiza contratos legales en español. Identifica: terminación unilateral, penalizaciones, jurisdicción, riesgos. Respuesta completa en español, sin emojis ni formato markdown.`
	systemPromptTokens := len(systemPrompt) / entities.CharsPerToken

	// User query tokens
	userQuery := `Analiza este contrato completo. Identifica terminación unilateral, penalizaciones (con montos), jurisdicción/arbitraje y riesgos principales. Respuesta completa en español, sin emojis, sin formato markdown.`
	userQueryTokens := len(userQuery) / entities.CharsPerToken

	// Total input tokens
	totalInputTokens := systemPromptTokens + userQueryTokens + estimatedInputTokens

	// Output tokens
	outputTokens := maxTokensConfig
	if outputTokens < 800 {
		outputTokens = 800
	}
	if outputTokens > entities.MaxOutputTokens {
		outputTokens = entities.MaxOutputTokens
	}

	totalTokens := totalInputTokens + outputTokens

	// Establecer valores en la estimación
	estimation.EstimatedTokens = estimatedInputTokens
	estimation.Chunks = 1 // Procesamiento en una sola petición
	estimation.SystemPromptTokens = systemPromptTokens
	estimation.Phase1Tokens = 0 // No hay fase 1
	estimation.Phase2InputTokens = totalInputTokens
	estimation.Phase2OutputTokens = outputTokens
	estimation.TotalTokens = totalTokens
	estimation.RecommendedMaxTokens = outputTokens

	return estimation
}

func (uc *EstimateTokensUseCase) calculateRecommendedMaxTokens(numChunks, charCount int) int {
	recommendedMaxTokens := 800
	if numChunks > 10 {
		recommendedMaxTokens = 1000
	}
	if numChunks > 20 {
		recommendedMaxTokens = 1500
	}
	if charCount > 50000 {
		recommendedMaxTokens = 2000
	}
	return recommendedMaxTokens
}

func (uc *EstimateTokensUseCase) generateWarning(numChunks, totalTokens int) string {
	if numChunks > 30 {
		return "Documento muy grande. El análisis puede tardar más de 20 minutos. Considera dividir el documento."
	} else if numChunks > 15 {
		return "Documento grande. El análisis puede tardar 10-20 minutos."
	} else if totalTokens > 100000 {
		return "Alto uso de tokens. Considera usar un modelo con mayor contexto o reducir maxTokens."
	}
	return ""
}
