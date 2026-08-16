// Package main is the entry point for the Meta Clash Go backend.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/auth"
	"github.com/nikhilsaxena04/meta_clash/backend/internal/config"
	"github.com/nikhilsaxena04/meta_clash/backend/internal/db"
	"github.com/nikhilsaxena04/meta_clash/backend/internal/game"
	"github.com/nikhilsaxena04/meta_clash/backend/internal/lobby"
	"github.com/nikhilsaxena04/meta_clash/backend/internal/middleware"
	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
	"github.com/nikhilsaxena04/meta_clash/backend/internal/ws"
)

func main() {
	// Structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()

	// ── Database ───────────────────────────────────────────────
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.Migrate(pool); err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	repo := db.NewPostgresRepo(pool)

	// ── Auth ───────────────────────────────────────────────────
	authHandler := auth.NewAuthHandler(repo, cfg.JWTSecret, cfg.JWTExpiry)

	// ── Game Dependencies ──────────────────────────────────────
	superheroClient := game.NewSuperheroClient(cfg.GeminiAPIKey)
	cardGen := game.NewGenerator(cfg.JikanBaseURL, cfg.JikanTimeout, cfg.GeminiAPIKey, superheroClient)
	engine := game.NewEngine()
	store := lobby.NewMemoryStore()
	manager := lobby.NewManager(store, cardGen, engine, repo)

	// ── WebSocket Hub ──────────────────────────────────────────
	hub := ws.NewHub()
	go hub.Run()
	wsHandlers := ws.NewHandlers(manager, hub)

	// ── HTTP Mux ───────────────────────────────────────────────
	mux := http.NewServeMux()

	// Auth endpoints
	mux.HandleFunc("POST /api/auth/register", authHandler.HandleRegister)
	mux.HandleFunc("POST /api/auth/login", authHandler.HandleLogin)

	// User profile endpoint
	mux.HandleFunc("GET /api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		if userID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing user ID"})
			return
		}

		user, err := repo.GetByID(userID)
		if err != nil {
			slog.Error("get user profile failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
			return
		}
		if user == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
			return
		}

		wins, losses, err := repo.GetWinLoss(userID)
		if err != nil {
			slog.Error("get win/loss failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
			return
		}

		history, err := repo.GetMatchHistory(userID, 20)
		if err != nil {
			slog.Error("get match history failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
			return
		}

		profile := models.UserProfile{
			User:    *user,
			Wins:    wins,
			Losses:  losses,
			History: history,
		}

		writeJSON(w, http.StatusOK, profile)
	})

	// WebSocket endpoint (with optional auth for guest mode)
	optAuth := auth.OptionalAuth(cfg.JWTSecret)
	mux.Handle("GET /api/ws", optAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, wsHandlers, w, r)
	})))

	// Health check
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().UTC().Format(time.RFC3339))
	})

	// Readiness check
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready"}`)
	})

	// Debug endpoint to demonstrate the game engine
	mux.HandleFunc("GET /api/debug/lobby", func(w http.ResponseWriter, r *http.Request) {
		debugManager := lobby.NewManager(store, cardGen, engine, repo)
		theme := r.URL.Query().Get("theme")
		if theme == "" {
			theme = "one piece"
		}

		host := models.Player{
			ID:   models.PlayerID("host-123"),
			Name: "HumanPlayer",
		}

		l, err := debugManager.CreateLobby(theme, host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Fill with bots and start
		l, err = debugManager.StartGame(l.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(l)
	})

	// ── Middleware Stack ────────────────────────────────────────
	// Applied outermost-first: Recovery → CORS → Logging → mux
	var handler http.Handler = mux
	handler = middleware.Logging(handler)
	handler = middleware.CORS(cfg.AllowedOrigin)(handler)
	handler = middleware.Recovery(handler)

	// ── Server ─────────────────────────────────────────────────
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		slog.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutting down", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

// writeJSON is a helper that encodes data as JSON and writes it to the response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
