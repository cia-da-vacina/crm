package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/module/auditlog"
	"github.com/cia-da-vacina/crm/backend/internal/module/auth"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation"
	"github.com/cia-da-vacina/crm/backend/internal/module/customer"
	"github.com/cia-da-vacina/crm/backend/internal/module/dashboard"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement"
	"github.com/cia-da-vacina/crm/backend/internal/module/followup"
	"github.com/cia-da-vacina/crm/backend/internal/module/lossreason"
	"github.com/cia-da-vacina/crm/backend/internal/module/me"
	"github.com/cia-da-vacina/crm/backend/internal/module/metasettings"
	"github.com/cia-da-vacina/crm/backend/internal/module/pop"
	"github.com/cia-da-vacina/crm/backend/internal/module/pricing"
	"github.com/cia-da-vacina/crm/backend/internal/module/template"
	"github.com/cia-da-vacina/crm/backend/internal/module/triage"
	"github.com/cia-da-vacina/crm/backend/internal/module/unit"
	"github.com/cia-da-vacina/crm/backend/internal/module/user"
	"github.com/cia-da-vacina/crm/backend/internal/module/webhook"
	"github.com/cia-da-vacina/crm/backend/pkg/env"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := env.LoadFile(".env"); err != nil {
		log.Printf("warning: failed to load .env: %v", err)
	}
	log.Println("starting CRM Cia da Vacina API...")

	a := app.New()
	defer a.Close()

	router := httppkg.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.RequestID)
	router.Use(httppkg.CORS)
	router.Use(httprate.LimitByIP(500, 1*time.Minute))
	router.Use(middleware.Logger)
	router.Use(httppkg.Recoverer)
	router.Use(middleware.ThrottleBacklog(100, 200, 2*time.Second))
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(middleware.Compress(5))

	registerRoutes(router, a)

	port := env.GetOrDefault("API_PORT", "8080")

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 65 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("stopped")
}

// registerRoutes monta as rotas da API. Os módulos de domínio entram aqui
// incrementalmente, na ordem descrita em backend/ARCHITECTURE.md §8.
// /health fica fora de /api/v1 (mesma convenção do docs/openapi.yaml).
func registerRoutes(r *httppkg.Router, a *app.App) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httppkg.Success(w, http.StatusOK, "ok", map[string]string{"status": "ok"})
	})

	// Fora de /api/v1 de propósito — a Meta chama isso direto, nunca o
	// frontend (docs/BACKEND-CONTRACT.md §8).
	webhook.New(a).Register(r)

	r.Group("/api/v1", func(r *httppkg.Router) {
		auth.New(a).Register(r)
		user.New(a).Register(r)
		unit.New(a).Register(r)
		me.New(a).Register(r)
		customer.New(a).Register(r)
		conversation.New(a).Register(r)
		followup.New(a).Register(r)
		pop.New(a).Register(r)
		lossreason.New(a).Register(r)
		dashboard.New(a).Register(r)
		metasettings.New(a).Register(r)
		triage.New(a).Register(r)
		engagement.New(a).Register(r)
		auditlog.New(a).Register(r)
		pricing.New(a).Register(r)
		template.New(a).Register(r)
	})
}
