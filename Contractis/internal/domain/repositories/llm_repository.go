package repositories

import (
	"context"
	"github.com/rodascaar/contractis/internal/domain/entities"
)

// LLMRepository define la interfaz para interactuar con modelos de lenguaje
type LLMRepository interface {
	// SendChatRequest envía una solicitud de chat al LLM
	SendChatRequest(ctx context.Context, config *entities.LLMConfig, messages []ChatMessage, maxTokens int) (string, error)

	// TestConnection verifica la conexión con el LLM
	TestConnection(ctx context.Context, config *entities.LLMConfig) error
}

// ChatMessage representa un mensaje en el chat
type ChatMessage struct {
	Role    string
	Content string
}
