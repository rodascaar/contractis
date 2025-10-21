package repositories

import (
	"context"
	"time"

	"github.com/rodascaar/contractis/internal/domain/entities"
)

// ContractRepository define la interfaz para persistencia de contratos
type ContractRepository interface {
	// Create crea un nuevo registro de contrato
	Create(ctx context.Context, record *entities.ContractRecord) (int64, error)

	// GetByID obtiene un contrato por su ID
	GetByID(ctx context.Context, id int64) (*entities.ContractRecord, error)

	// GetByHash obtiene un contrato por su hash (para caché)
	GetByHash(ctx context.Context, hash string) (*entities.ContractRecord, error)

	// Update actualiza un registro de contrato
	Update(ctx context.Context, record *entities.ContractRecord) error

	// List lista todos los contratos con paginación
	List(ctx context.Context, limit, offset int) ([]*entities.ContractRecord, error)

	// Search busca contratos por nombre de archivo
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.ContractRecord, error)

	// GetStats obtiene estadísticas de contratos
	GetStats(ctx context.Context) (*ContractStats, error)

	// Delete elimina un contrato por ID
	Delete(ctx context.Context, id int64) error

	// GetRecent obtiene los contratos más recientes
	GetRecent(ctx context.Context, limit int) ([]*entities.ContractRecord, error)
}

// ContractStats representa estadísticas de contratos
type ContractStats struct {
	TotalContracts        int        `json:"TotalContracts"`
	CompletedContracts    int        `json:"CompletedContracts"`
	FailedContracts       int        `json:"FailedContracts"`
	TotalProcessingTime   float64    `json:"TotalProcessingTime"`
	AverageProcessingTime float64    `json:"AverageProcessingTime"`
	LastAnalyzedAt        *time.Time `json:"LastAnalyzedAt,omitempty"`
}
