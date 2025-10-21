package entities

import "time"

// AnalysisResult representa el resultado del análisis de un contrato
type AnalysisResult struct {
	Success     bool
	ContractID  string
	Content     string
	Error       string
	ProcessedAt time.Time
	DurationSec float64
	TokensUsed  int
	ChunksCount int
}

// NewAnalysisResult crea un nuevo resultado de análisis
func NewAnalysisResult(contractID string) *AnalysisResult {
	return &AnalysisResult{
		ContractID:  contractID,
		ProcessedAt: time.Now(),
	}
}

// MarkSuccess marca el análisis como exitoso
func (a *AnalysisResult) MarkSuccess(content string, duration time.Duration, chunks int) {
	a.Success = true
	a.Content = content
	a.DurationSec = duration.Seconds()
	a.ChunksCount = chunks
}

// MarkFailure marca el análisis como fallido
func (a *AnalysisResult) MarkFailure(err error) {
	a.Success = false
	a.Error = err.Error()
}

// TokenEstimation representa la estimación de tokens para un análisis
type TokenEstimation struct {
	Success              bool
	CharacterCount       int
	EstimatedTokens      int
	Chunks               int
	SystemPromptTokens   int
	Phase1Tokens         int
	Phase2InputTokens    int
	Phase2OutputTokens   int
	TotalTokens          int
	RecommendedMaxTokens int
	Warning              string
	Error                string
}

// NewTokenEstimation crea una nueva estimación de tokens
func NewTokenEstimation(charCount int) *TokenEstimation {
	return &TokenEstimation{
		Success:        true,
		CharacterCount: charCount,
	}
}

// SetWarning establece una advertencia en la estimación
func (t *TokenEstimation) SetWarning(warning string) {
	t.Warning = warning
}

// SetError marca la estimación como fallida
func (t *TokenEstimation) SetError(err error) {
	t.Success = false
	t.Error = err.Error()
}
