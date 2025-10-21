package dto

// LLMConfigRequest representa la configuración del LLM en las peticiones HTTP
type LLMConfigRequest struct {
	Type      string `json:"type"`
	LocalUrl  string `json:"localUrl"`
	ApiUrl    string `json:"apiUrl"`
	ApiKey    string `json:"apiKey"`
	ModelName string `json:"modelName"`
	MaxTokens int    `json:"maxTokens"`
}
