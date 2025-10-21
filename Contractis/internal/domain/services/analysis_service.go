package services

import (
	"context"
	"github.com/rodascaar/contractis/internal/domain/entities"
)

// AnalysisService define la interfaz para el servicio de análisis
type AnalysisService interface {
	// AnalyzeContract analiza un contrato completo
	AnalyzeContract(ctx context.Context, contract *entities.Contract, config *entities.LLMConfig) (*entities.AnalysisResult, error)

	// EstimateTokenUsage estima el uso de tokens para un contrato
	EstimateTokenUsage(contract *entities.Contract, maxTokens int) (*entities.TokenEstimation, error)
}
