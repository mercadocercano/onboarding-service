package migration

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migrator maneja las migraciones de base de datos
type Migrator struct {
	db             *sql.DB
	migrationsPath string
}

// NewMigrator crea una nueva instancia del migrator
func NewMigrator(db *sql.DB, migrationsPath string) *Migrator {
	return &Migrator{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

// RunMigrations ejecuta todas las migraciones pendientes
func (m *Migrator) RunMigrations() error {
	log.Println("🚀 Starting database migrations...")

	// Crear tabla de migraciones si no existe
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	// Obtener migraciones ejecutadas
	executedMigrations, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("error getting executed migrations: %w", err)
	}

	// Obtener archivos de migración
	migrationFiles, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("error getting migration files: %w", err)
	}

	// Ejecutar migraciones pendientes
	for _, file := range migrationFiles {
		migrationName := strings.TrimSuffix(file, ".sql")
		
		if !contains(executedMigrations, migrationName) {
			log.Printf("📝 Executing migration: %s", migrationName)
			
			if err := m.executeMigration(file); err != nil {
				return fmt.Errorf("error executing migration %s: %w", file, err)
			}
			
			if err := m.recordMigration(migrationName); err != nil {
				return fmt.Errorf("error recording migration %s: %w", migrationName, err)
			}
			
			log.Printf("✅ Migration %s completed successfully", migrationName)
		} else {
			log.Printf("⏭️  Skipping already executed migration: %s", migrationName)
		}
	}

	log.Printf("🎉 All migrations completed successfully!")
	return nil
}

// createMigrationsTable crea la tabla para trackear migraciones
func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			migration_name VARCHAR(255) UNIQUE NOT NULL,
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`
	
	_, err := m.db.Exec(query)
	return err
}

// getExecutedMigrations obtiene la lista de migraciones ya ejecutadas
func (m *Migrator) getExecutedMigrations() ([]string, error) {
	query := `SELECT migration_name FROM schema_migrations ORDER BY id`
	
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		migrations = append(migrations, name)
	}

	return migrations, nil
}

// getMigrationFiles obtiene la lista de archivos de migración ordenados
func (m *Migrator) getMigrationFiles() ([]string, error) {
	files, err := os.ReadDir(m.migrationsPath)
	if err != nil {
		return nil, err
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// Ordenar archivos para ejecución secuencial
	sort.Strings(migrationFiles)
	
	return migrationFiles, nil
}

// executeMigration ejecuta un archivo de migración específico
func (m *Migrator) executeMigration(filename string) error {
	filePath := filepath.Join(m.migrationsPath, filename)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Dividir respetando bloques $$ (PL/pgSQL)
	statements := splitSQLStatements(string(content))

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" || strings.HasPrefix(statement, "--") {
			continue
		}

		if _, err := tx.Exec(statement); err != nil {
			return fmt.Errorf("error executing statement '%s': %w", truncateForError(statement, 80), err)
		}
	}

	return tx.Commit()
}

// splitSQLStatements divide SQL respetando bloques $$...$$
func splitSQLStatements(content string) []string {
	var statements []string
	var current strings.Builder
	inDollarQuote := false

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Saltar líneas que son solo comentarios
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			if !inDollarQuote {
				continue
			}
		}

		// Detectar inicio/fin de bloque $$ (cada $$ alterna el estado)
		for i := 0; i < strings.Count(line, "$$"); i++ {
			inDollarQuote = !inDollarQuote
		}

		current.WriteString(line)
		current.WriteString("\n")

		// Solo dividir por ; cuando NO estamos dentro de $$
		if !inDollarQuote && strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	if current.Len() > 0 {
		stmt := strings.TrimSpace(current.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}

func truncateForError(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// recordMigration registra una migración como ejecutada
func (m *Migrator) recordMigration(migrationName string) error {
	query := `INSERT INTO schema_migrations (migration_name) VALUES ($1)`
	_, err := m.db.Exec(query, migrationName)
	return err
}

// contains verifica si un slice contiene un elemento
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetMigrationsPath resuelve la ruta de migraciones relativa al directorio de trabajo
func GetMigrationsPath() string {
	// Intentar encontrar la ruta de migraciones
	possiblePaths := []string{
		"migrations",
		"../migrations", 
		"../../migrations",
		"../../../migrations",
		"./services/saas-mt-onboarding-service/migrations",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			abs, _ := filepath.Abs(path)
			return abs
		}
	}

	// Fallback
	log.Println("⚠️  Warning: Could not find migrations directory, using default")
	return "migrations"
} 