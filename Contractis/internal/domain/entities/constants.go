package entities

import "time"

const (
	// File constraints
	MaxFileSize = 10 * 1024 * 1024 // 10MB

	// Model constraints (basado en Qwen 3 4B)
	QwenContextWindow   = 8000
	OnlineContextWindow = 128000 // Para modelos online como GPT-4
	SafetyMargin        = 1000
	MaxOutputTokens     = 2000
	MaxInputTokens      = QwenContextWindow - MaxOutputTokens - SafetyMargin

	// Chunk configuration
	DefaultChunkSize = 3000
	Phase1MaxTokens  = 1200
	MinChunkSize     = 500
	CharsPerToken    = 3
	MaxFragments     = 4

	// Timeout configuration
	HTTPTimeoutLocal      = 120 * time.Second // Para modelos locales (pueden ser lentos)
	HTTPTimeoutOnline     = 5 * time.Minute   // Para modelos online (necesitan tiempo para procesar chunks grandes)
	TestConnectionTimeout = 10 * time.Second  // Para pruebas de conexión
	ConsolidationTimeout  = 10 * time.Minute

	// Retry configuration
	MaxRetries = 3
	RetryDelay = 2 * time.Second
)
