// Package app holds the shared infrastructure passed to all modules.
package app

import (
	"log"

	"github.com/cia-da-vacina/crm/backend/pkg/audit"
	"github.com/cia-da-vacina/crm/backend/pkg/crypto"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
	"github.com/cia-da-vacina/crm/backend/pkg/env"
	"github.com/cia-da-vacina/crm/backend/pkg/jwt"
	"github.com/cia-da-vacina/crm/backend/pkg/meta"
	"github.com/cia-da-vacina/crm/backend/pkg/sse"
)

// App holds every shared infrastructure dependency.
// Modules receive *App and create their own internal dependencies from it.
type App struct {
	DB     *database.DB
	JWT    *jwt.Service
	Meta   *meta.Registry
	SSE    *sse.Hub
	Crypto *crypto.Cipher
	Audit  *audit.Logger
}

func New() *App {
	db := mustConnectDB()
	jwtSvc := mustNewJWT()

	return &App{
		DB:     db,
		JWT:    jwtSvc,
		Meta:   newMetaRegistry(),
		SSE:    sse.NewHub(),
		Crypto: mustNewCipher(),
		Audit:  audit.New(db),
	}
}

func (a *App) Close() {
	if err := a.DB.Close(); err != nil {
		log.Printf("error closing DB: %v", err)
	}
}

func mustConnectDB() *database.DB {
	dsn := env.GetOrDefault("DATABASE_URL", "")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	db, err := database.NewDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("database connected")
	return db
}

func mustNewJWT() *jwt.Service {
	svc, err := jwt.NewServiceFromEnv()
	if err != nil {
		log.Fatalf("failed to create JWT service: %v", err)
	}
	log.Println("JWT service ready")
	return svc
}

func mustNewCipher() *crypto.Cipher {
	c, err := crypto.NewCipherFromEnv()
	if err != nil {
		log.Fatalf("failed to create cipher: %v", err)
	}
	log.Println("crypto cipher ready")
	return c
}

// newMetaRegistry registra MockClient pros três canais — nenhuma chamada de
// rede real acontece ainda. Não existem credenciais Meta reais neste
// ambiente de dev (ver backend/ARCHITECTURE.md §6): os clients HTTP reais
// (pkg/meta/whatsapp.go, instagram.go) já existem prontos pra registrar aqui
// no lugar do mock assim que houver token de produção — nenhum usecase muda.
func newMetaRegistry() *meta.Registry {
	registry := meta.NewRegistry()
	registry.Register(meta.NewMockClient(meta.ChannelWhatsApp))
	registry.Register(meta.NewMockClient(meta.ChannelInstagram))
	registry.Register(meta.NewMockClient(meta.ChannelFacebook))
	log.Println("meta registry ready (mock clients — no real Meta calls yet)")
	return registry
}
