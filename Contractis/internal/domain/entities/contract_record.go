package entities

import "time"

// ContractStatus representa el estado de un análisis de contrato
type ContractStatus string

const (
	StatusPending   ContractStatus = "pending"
	StatusAnalyzing ContractStatus = "analyzing"
	StatusCompleted ContractStatus = "completed"
	StatusFailed    ContractStatus = "failed"
)

// ContractRecord representa un registro de contrato analizado en la base de datos
type ContractRecord struct {
	ID         int64          `json:"id"`
	Filename   string         `json:"filename"`
	FileHash   string         `json:"file_hash"`
	FileSize   int64          `json:"file_size"`
	UploadedAt time.Time      `json:"uploaded_at"`
	AnalyzedAt *time.Time     `json:"analyzed_at,omitempty"`
	Status     ContractStatus `json:"status"`

	// Metadata del análisis
	LLMType   string `json:"llm_type"`
	LLMModel  string `json:"llm_model"`
	MaxTokens int    `json:"max_tokens"`

	// Resultados
	AnalysisResult        string  `json:"analysis_result"`
	CharacterCount        int     `json:"character_count"`
	EstimatedTokens       int     `json:"estimated_tokens"`
	ChunksCount           int     `json:"chunks_count"`
	ProcessingTimeSeconds float64 `json:"processing_time_seconds"`

	// Trazabilidad
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewContractRecord crea un nuevo registro de contrato
func NewContractRecord(filename, fileHash string, fileSize int64, llmType, llmModel string, maxTokens int) *ContractRecord {
	now := time.Now()
	return &ContractRecord{
		Filename:   filename,
		FileHash:   fileHash,
		FileSize:   fileSize,
		UploadedAt: now,
		Status:     StatusPending,
		LLMType:    llmType,
		LLMModel:   llmModel,
		MaxTokens:  maxTokens,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// MarkAnalyzing marca el contrato como en análisis
func (cr *ContractRecord) MarkAnalyzing() {
	cr.Status = StatusAnalyzing
	cr.UpdatedAt = time.Now()
}

// MarkCompleted marca el contrato como completado
func (cr *ContractRecord) MarkCompleted(result string, charCount, tokens, chunks int, processingTime float64) {
	now := time.Now()
	cr.Status = StatusCompleted
	cr.AnalysisResult = result
	cr.CharacterCount = charCount
	cr.EstimatedTokens = tokens
	cr.ChunksCount = chunks
	cr.ProcessingTimeSeconds = processingTime
	cr.AnalyzedAt = &now
	cr.UpdatedAt = now
}

// MarkFailed marca el contrato como fallido
func (cr *ContractRecord) MarkFailed(errorMsg string) {
	cr.Status = StatusFailed
	cr.ErrorMessage = errorMsg
	cr.UpdatedAt = time.Now()
}
