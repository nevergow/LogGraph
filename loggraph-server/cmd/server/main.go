package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"loggraph/internal/config"
	"loggraph/internal/database"
	"loggraph/internal/handler"
	"loggraph/internal/repository"
	"loggraph/internal/router"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connect: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	// Repositories
	blockRepo := repository.NewBlockRepo(pool)
	nodeRepo := repository.NewNodeRepo(pool)
	relationRepo := repository.NewRelationRepo(pool)
	webhookTokenRepo := repository.NewWebhookTokenRepo(pool)

	// Handlers
	blockHandler := handler.NewBlockHandler(blockRepo, relationRepo)
	nodeHandler := handler.NewNodeHandler(nodeRepo)
	webhookHandler := handler.NewWebhookHandler(blockRepo, webhookTokenRepo)
	larkHandler := handler.NewLarkHandler(blockRepo, webhookTokenRepo)
	attachmentHandler := handler.NewAttachmentHandler(cfg)

	r := router.New(blockHandler, nodeHandler, webhookHandler, larkHandler, attachmentHandler)

	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[server] listening on %s", cfg.Addr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[server] shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("[server] stopped")
}
