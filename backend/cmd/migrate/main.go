package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cia-da-vacina/crm/backend/pkg/env"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if err := env.LoadFile(".env"); err != nil {
		log.Printf("warning: failed to load .env: %v", err)
	}
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Println("No database URL provided")
		return
	}

	// golang-migrate (driver pgx/v5) usa o esquema "pgx5"; o DATABASE_URL do app usa "postgres".
	migrateURL := dbURL
	if strings.HasPrefix(migrateURL, "postgres://") {
		migrateURL = "pgx5://" + strings.TrimPrefix(migrateURL, "postgres://")
	} else if strings.HasPrefix(migrateURL, "postgresql://") {
		migrateURL = "pgx5://" + strings.TrimPrefix(migrateURL, "postgresql://")
	}

	source := "file://./conf/migrations"

	m, err := migrate.New(source, migrateURL)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	cmd := flag.Arg(0)
	if cmd == "" {
		cmd = "up"
	}

	switch cmd {
	case "up":
		if err := m.Up(); err != nil {
			if err == migrate.ErrNoChange {
				fmt.Println("No migration to apply")
				return
			}
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil {
			if err == migrate.ErrNoChange {
				fmt.Println("Already at oldest version")
				return
			}
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
		fmt.Println("Migration rolled back successfully")

	case "force":
		version := flag.Arg(1)
		if version == "" {
			log.Fatal("Version required for force command")
		}
		var v int64
		fmt.Sscanf(version, "%d", &v)
		if err := m.Force(int(v)); err != nil {
			log.Fatalf("Failed to force version: %v", err)
		}
		fmt.Printf("Forced to version %d\n", v)

	case "drop":
		if err := m.Drop(); err != nil {
			log.Fatalf("Failed to drop migrations: %v", err)
		}
		fmt.Println("Migrations dropped successfully")

	default:
		log.Fatalf("Unknown command: %s", cmd)
	}
}
