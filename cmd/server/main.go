package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"chat-backend/internal/config"
	"chat-backend/internal/pubsub"
	"chat-backend/internal/storage"
	"chat-backend/internal/ws"
)

func main() {
	cfg := config.Load()

	store, err := storage.NewPostgres(cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName)
	if err != nil {
		log.Fatalf("postgres init: %v", err)
	}
	defer store.Close()

	ps := pubsub.NewRedis(cfg.RedisAddr)
	defer ps.Close()

	hub := ws.NewHub(128)
	handler := ws.NewHandler(hub, store, ps, cfg.FrontendOrigin)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go hub.Run(ctx)
	if err := handler.StartRedisFanIn(ctx); err != nil {
		log.Fatalf("redis sub: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handler.WebSocket)
	mux.HandleFunc("/messages", handler.History)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Listening on :%s (HTTPS/WSS)", cfg.Port)
		if err := srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
