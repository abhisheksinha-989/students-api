package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abhisheksinha-989/students-api/internal/config"
	"github.com/abhisheksinha-989/students-api/internal/http/handlers/student"
	"github.com/abhisheksinha-989/students-api/internal/logger"
	"github.com/abhisheksinha-989/students-api/internal/storage/postgres"
)

func main() {
	cfg := config.MustLoad()
	logger.Setup(cfg.Env)

	db, err := postgres.New(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Could not connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database connected", slog.String("env", cfg.Env))

	router := http.NewServeMux()
	router.HandleFunc("POST /api/students", student.Create(db))
	router.HandleFunc("GET /api/students", student.GetList(db))
	router.HandleFunc("GET /api/students/{id}", student.GetById(db))
	router.HandleFunc("PUT /api/students/{id}", student.Update(db))
	router.HandleFunc("DELETE /api/students/{id}", student.Delete(db))

	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server started", slog.String("address", cfg.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("shutting down server.....")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown gracefully", slog.String("error", err.Error()))
	}
	slog.Info("server stopped")
}
