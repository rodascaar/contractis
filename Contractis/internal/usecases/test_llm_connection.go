package usecases

import (
	"context"
	"github.com/rodascaar/contractis/internal/domain/entities"
	"github.com/rodascaar/contractis/internal/domain/repositories"
)

// TestLLMConnectionUseCase maneja el caso de uso de prueba de conexión con LLM
type TestLLMConnectionUseCase struct {
	llmRepo repositories.LLMRepository
}

// NewTestLLMConnectionUseCase crea una nueva instancia del caso de uso
func NewTestLLMConnectionUseCase(llmRepo repositories.LLMRepository) *TestLLMConnectionUseCase {
	return &TestLLMConnectionUseCase{
		llmRepo: llmRepo,
	}
}

// Execute ejecuta la prueba de conexión
func (uc *TestLLMConnectionUseCase) Execute(ctx context.Context, config *entities.LLMConfig) error {
	// Validar configuración
	if err := config.Validate(); err != nil {
		return err
	}

	// Probar conexión
	return uc.llmRepo.TestConnection(ctx, config)
}
