package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rodascaar/contractis/internal/domain/entities"
	"github.com/rodascaar/contractis/internal/domain/repositories"
)

// ContractRepositoryImpl implementa ContractRepository usando SQLite
type ContractRepositoryImpl struct {
	db *DB
}

// NewContractRepository crea una nueva instancia del repositorio
func NewContractRepository(db *DB) repositories.ContractRepository {
	return &ContractRepositoryImpl{db: db}
}

// Create crea un nuevo registro de contrato
func (r *ContractRepositoryImpl) Create(ctx context.Context, record *entities.ContractRecord) (int64, error) {
	query := `
		INSERT INTO contracts (
			filename, file_hash, file_size, uploaded_at, status,
			llm_type, llm_model, max_tokens, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		record.Filename,
		record.FileHash,
		record.FileSize,
		record.UploadedAt,
		record.Status,
		record.LLMType,
		record.LLMModel,
		record.MaxTokens,
		record.CreatedAt,
		record.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("error creating contract record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert id: %w", err)
	}

	return id, nil
}

// GetByID obtiene un contrato por su ID
func (r *ContractRepositoryImpl) GetByID(ctx context.Context, id int64) (*entities.ContractRecord, error) {
	query := `
		SELECT id, filename, file_hash, file_size, uploaded_at, analyzed_at, status,
		       llm_type, llm_model, max_tokens, analysis_result, character_count,
		       estimated_tokens, chunks_count, processing_time_seconds, error_message,
		       created_at, updated_at
		FROM contracts
		WHERE id = ?
	`

	record := &entities.ContractRecord{}
	var analyzedAt, uploadedAt, createdAt, updatedAt sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.Filename,
		&record.FileHash,
		&record.FileSize,
		&uploadedAt,
		&analyzedAt,
		&record.Status,
		&record.LLMType,
		&record.LLMModel,
		&record.MaxTokens,
		&record.AnalysisResult,
		&record.CharacterCount,
		&record.EstimatedTokens,
		&record.ChunksCount,
		&record.ProcessingTimeSeconds,
		&record.ErrorMessage,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("contract not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting contract: %w", err)
	}

	// Parse datetime strings
	if uploadedAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", uploadedAt.String); err == nil {
			record.UploadedAt = t
		}
	}
	if analyzedAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", analyzedAt.String); err == nil {
			record.AnalyzedAt = &t
		}
	}
	if createdAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", createdAt.String); err == nil {
			record.CreatedAt = t
		}
	}
	if updatedAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", updatedAt.String); err == nil {
			record.UpdatedAt = t
		}
	}

	return record, nil
}

// GetByHash obtiene un contrato por su hash (cualquier status)
func (r *ContractRepositoryImpl) GetByHash(ctx context.Context, hash string) (*entities.ContractRecord, error) {
	query := `
		SELECT id, filename, file_hash, file_size, uploaded_at, analyzed_at, status,
		       llm_type, llm_model, max_tokens, analysis_result, character_count,
		       estimated_tokens, chunks_count, processing_time_seconds, error_message,
		       created_at, updated_at
		FROM contracts
		WHERE file_hash = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	record := &entities.ContractRecord{}
	var analyzedAt, uploadedAt, createdAt, updatedAt sql.NullString

	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&record.ID,
		&record.Filename,
		&record.FileHash,
		&record.FileSize,
		&uploadedAt,
		&analyzedAt,
		&record.Status,
		&record.LLMType,
		&record.LLMModel,
		&record.MaxTokens,
		&record.AnalysisResult,
		&record.CharacterCount,
		&record.EstimatedTokens,
		&record.ChunksCount,
		&record.ProcessingTimeSeconds,
		&record.ErrorMessage,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No encontrado, no es error
	}
	if err != nil {
		return nil, fmt.Errorf("error getting contract by hash: %w", err)
	}

	// Parse datetime strings
	if uploadedAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", uploadedAt.String); err == nil {
			record.UploadedAt = t
		}
	}
	if analyzedAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", analyzedAt.String); err == nil {
			record.AnalyzedAt = &t
		}
	}
	if createdAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", createdAt.String); err == nil {
			record.CreatedAt = t
		}
	}
	if updatedAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", updatedAt.String); err == nil {
			record.UpdatedAt = t
		}
	}

	return record, nil
}

// Update actualiza un registro de contrato
func (r *ContractRepositoryImpl) Update(ctx context.Context, record *entities.ContractRecord) error {
	query := `
		UPDATE contracts SET
			status = ?,
			analyzed_at = ?,
			analysis_result = ?,
			character_count = ?,
			estimated_tokens = ?,
			chunks_count = ?,
			processing_time_seconds = ?,
			error_message = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		record.Status,
		record.AnalyzedAt,
		record.AnalysisResult,
		record.CharacterCount,
		record.EstimatedTokens,
		record.ChunksCount,
		record.ProcessingTimeSeconds,
		record.ErrorMessage,
		time.Now(),
		record.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating contract: %w", err)
	}

	return nil
}

// List lista todos los contratos con paginación
func (r *ContractRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.ContractRecord, error) {
	query := `
		SELECT id, filename, file_hash, file_size, uploaded_at, analyzed_at, status,
		       llm_type, llm_model, max_tokens, analysis_result, character_count,
		       estimated_tokens, chunks_count, processing_time_seconds, error_message,
		       created_at, updated_at
		FROM contracts
		ORDER BY uploaded_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing contracts: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// Search busca contratos por nombre de archivo
func (r *ContractRepositoryImpl) Search(ctx context.Context, query string, limit, offset int) ([]*entities.ContractRecord, error) {
	sqlQuery := `
		SELECT id, filename, file_hash, file_size, uploaded_at, analyzed_at, status,
		       llm_type, llm_model, max_tokens, analysis_result, character_count,
		       estimated_tokens, chunks_count, processing_time_seconds, error_message,
		       created_at, updated_at
		FROM contracts
		WHERE filename LIKE ?
		ORDER BY uploaded_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, "%"+query+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error searching contracts: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// GetRecent obtiene los contratos más recientes
func (r *ContractRepositoryImpl) GetRecent(ctx context.Context, limit int) ([]*entities.ContractRecord, error) {
	query := `
		SELECT id, filename, file_hash, file_size, uploaded_at, analyzed_at, status,
		       llm_type, llm_model, max_tokens, analysis_result, character_count,
		       estimated_tokens, chunks_count, processing_time_seconds, error_message,
		       created_at, updated_at
		FROM contracts
		WHERE status = 'completed'
		ORDER BY analyzed_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting recent contracts: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// GetStats obtiene estadísticas de contratos
func (r *ContractRepositoryImpl) GetStats(ctx context.Context) (*repositories.ContractStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0) as completed,
			COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) as failed,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN processing_time_seconds ELSE 0 END), 0.0) as total_time,
			COALESCE(AVG(CASE WHEN status = 'completed' THEN processing_time_seconds ELSE NULL END), 0.0) as avg_time,
			MAX(analyzed_at) as last_analyzed
		FROM contracts
	`

	stats := &repositories.ContractStats{}
	var lastAnalyzed sql.NullString

	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalContracts,
		&stats.CompletedContracts,
		&stats.FailedContracts,
		&stats.TotalProcessingTime,
		&stats.AverageProcessingTime,
		&lastAnalyzed,
	)

	if err != nil {
		return nil, fmt.Errorf("error getting stats: %w", err)
	}

	// Parse last_analyzed datetime string
	if lastAnalyzed.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", lastAnalyzed.String); err == nil {
			stats.LastAnalyzedAt = &t
		}
	}

	return stats, nil
}

// Delete elimina un contrato por ID
func (r *ContractRepositoryImpl) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM contracts WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting contract: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("contract with ID %d not found", id)
	}

	return nil
}

// scanRows es un helper para escanear múltiples filas
func (r *ContractRepositoryImpl) scanRows(rows *sql.Rows) ([]*entities.ContractRecord, error) {
	var records []*entities.ContractRecord

	for rows.Next() {
		record := &entities.ContractRecord{}
		var analyzedAt, uploadedAt, createdAt, updatedAt sql.NullString

		err := rows.Scan(
			&record.ID,
			&record.Filename,
			&record.FileHash,
			&record.FileSize,
			&uploadedAt,
			&analyzedAt,
			&record.Status,
			&record.LLMType,
			&record.LLMModel,
			&record.MaxTokens,
			&record.AnalysisResult,
			&record.CharacterCount,
			&record.EstimatedTokens,
			&record.ChunksCount,
			&record.ProcessingTimeSeconds,
			&record.ErrorMessage,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Parse datetime strings
		if uploadedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", uploadedAt.String); err == nil {
				record.UploadedAt = t
			}
		}
		if analyzedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", analyzedAt.String); err == nil {
				record.AnalyzedAt = &t
			}
		}
		if createdAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", createdAt.String); err == nil {
				record.CreatedAt = t
			}
		}
		if updatedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", updatedAt.String); err == nil {
				record.UpdatedAt = t
			}
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return records, nil
}
