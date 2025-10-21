package dto

// AnalysisResponse representa la respuesta del análisis
type AnalysisResponse struct {
	Success bool   `json:"success"`
	Data    string `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// TokenEstimationResponse representa la respuesta de estimación de tokens
type TokenEstimationResponse struct {
	Success              bool   `json:"success"`
	CharacterCount       int    `json:"characterCount"`
	EstimatedTokens      int    `json:"estimatedTokens"`
	Chunks               int    `json:"chunks"`
	SystemPromptTokens   int    `json:"systemPromptTokens"`
	Phase1Tokens         int    `json:"phase1Tokens"`
	Phase2InputTokens    int    `json:"phase2InputTokens"`
	Phase2OutputTokens   int    `json:"phase2OutputTokens"`
	TotalTokens          int    `json:"totalTokens"`
	RecommendedMaxTokens int    `json:"recommendedMaxTokens"`
	Warning              string `json:"warning,omitempty"`
	Error                string `json:"error,omitempty"`
}
