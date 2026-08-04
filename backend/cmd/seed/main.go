package main

import (
	"log"
	"os"

	"github.com/cia-da-vacina/crm/backend/internal/seeder"
	envpkg "github.com/cia-da-vacina/crm/backend/pkg/env"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	if err := envpkg.LoadFile(".env"); err != nil {
		log.Printf("warning: failed to load .env: %v", err)
	}

	env := "prod"
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	if env != "dev" && env != "prod" {
		log.Fatalf("Unknown environment %q. Use: seed dev | seed prod", env)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	var seeders []seeder.Seeder

	switch env {
	case "prod":
		log.Println("Running production seeds (essential data)...")
		seeders = seeder.EssentialSeeders()
	case "dev":
		log.Println("Running dev seeds (essential + demo data)...")
		seeders = append(seeder.EssentialSeeders(), seeder.DemoSeeders()...)
	}

	if err := seeder.Run(db, seeders...); err != nil {
		log.Fatalf("Failed to run seeds: %v", err)
	}

	log.Printf("%s seeds applied successfully.", env)
}
