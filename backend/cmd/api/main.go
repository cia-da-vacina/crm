package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/cia-vacina/crm/backend/internal/auth"
	"github.com/cia-vacina/crm/backend/internal/config"
	"github.com/cia-vacina/crm/backend/internal/crmdemo"
	"github.com/cia-vacina/crm/backend/internal/httpapi"
	"github.com/cia-vacina/crm/backend/internal/store/memory"
)

func main() {
	cfg := config.Load()
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	st := memory.New()
	units, _ := st.ListUnits(context.Background())
	crm := crmdemo.New(units)
	tokens := auth.NewTokenService(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	srv := &httpapi.Server{Store: st, Tokens: tokens, CRM: crm}
	handler := srv.Routes()

	slog.Info("crm api starting", "addr", cfg.Addr, "env", cfg.Env)
	if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
