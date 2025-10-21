package entities

import "errors"

// LLMConfig representa la configuración del modelo de lenguaje
type LLMConfig struct {
	Type      string
	LocalUrl  string
	ApiUrl    string
	ApiKey    string
	ModelName string
	MaxTokens int
}

// NewLLMConfig crea una nueva configuración de LLM
func NewLLMConfig(llmType, localUrl, apiUrl, apiKey, modelName string, maxTokens int) *LLMConfig {
	return &LLMConfig{
		Type:      llmType,
		LocalUrl:  localUrl,
		ApiUrl:    apiUrl,
		ApiKey:    apiKey,
		ModelName: modelName,
		MaxTokens: maxTokens,
	}
}

// Validate valida la configuración del LLM
func (cfg *LLMConfig) Validate() error {
	if cfg.Type != "local" && cfg.Type != "online" {
		return errors.New("type must be 'local' or 'online'")
	}
	if cfg.Type == "online" {
		if cfg.ApiUrl == "" {
			return errors.New("apiUrl is required for online LLM")
		}
		if cfg.ApiKey == "" {
			return errors.New("apiKey is required for online LLM")
		}
		if cfg.ModelName == "" {
			return errors.New("modelName is required for online LLM")
		}
	}
	if cfg.Type == "local" && cfg.LocalUrl == "" {
		return errors.New("localUrl is required for local LLM")
	}
	if cfg.MaxTokens <= 0 {
		return errors.New("maxTokens must be positive")
	}
	return nil
}

// GetEndpointURL retorna la URL del endpoint según el tipo de LLM
func (cfg *LLMConfig) GetEndpointURL() string {
	if cfg.Type == "local" {
		return cfg.LocalUrl
	}
	return cfg.ApiUrl
}

// IsOnline verifica si el LLM está configurado para uso online
func (cfg *LLMConfig) IsOnline() bool {
	return cfg.Type == "online"
}

// GetAuthorizationHeader retorna el header de autorización si es necesario
func (cfg *LLMConfig) GetAuthorizationHeader() string {
	if cfg.IsOnline() && cfg.ApiKey != "" {
		return "Bearer " + cfg.ApiKey
	}
	return ""
}
