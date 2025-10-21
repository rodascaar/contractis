package database

const CreateContractsTableSQL = `
CREATE TABLE IF NOT EXISTS contracts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT NOT NULL,
    file_hash TEXT UNIQUE NOT NULL,
    file_size INTEGER NOT NULL,
    uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    analyzed_at DATETIME,
    status TEXT CHECK(status IN ('pending', 'analyzing', 'completed', 'failed')) DEFAULT 'pending',
    
    -- Metadata del análisis
    llm_type TEXT,
    llm_model TEXT,
    max_tokens INTEGER,
    
    -- Resultados
    analysis_result TEXT,
    character_count INTEGER,
    estimated_tokens INTEGER,
    chunks_count INTEGER,
    processing_time_seconds REAL,
    
    -- Trazabilidad
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_contracts_filename ON contracts(filename);
CREATE INDEX IF NOT EXISTS idx_contracts_uploaded_at ON contracts(uploaded_at);
CREATE INDEX IF NOT EXISTS idx_contracts_status ON contracts(status);
CREATE INDEX IF NOT EXISTS idx_contracts_file_hash ON contracts(file_hash);
CREATE INDEX IF NOT EXISTS idx_contracts_analyzed_at ON contracts(analyzed_at);
`

// RunMigrations ejecuta todas las migraciones necesarias
func RunMigrations(db *DB) error {
	_, err := db.Exec(CreateContractsTableSQL)
	if err != nil {
		return err
	}
	return nil
}
