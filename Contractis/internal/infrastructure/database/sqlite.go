package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

// DB envuelve la conexión a SQLite
type DB struct {
	*sql.DB
}

// NewSQLiteDB crea una nueva conexión a SQLite
func NewSQLiteDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Configurar conexión
	db.SetMaxOpenConns(1) // SQLite funciona mejor con una sola conexión

	// Verificar conexión
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	wrapper := &DB{DB: db}

	// Ejecutar migraciones
	if err := RunMigrations(wrapper); err != nil {
		return nil, fmt.Errorf("error running migrations: %w", err)
	}

	log.Printf("✓ Base de datos SQLite inicializada: %s", dbPath)
	return wrapper, nil
}

// Close cierra la conexión a la base de datos
func (db *DB) Close() error {
	return db.DB.Close()
}
