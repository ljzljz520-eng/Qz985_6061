package main

import (
	"context"
	"errors"
	"example.com/knowledge-backend/api"
	"example.com/knowledge-backend/config"
	"example.com/knowledge-backend/service"
	"example.com/knowledge-backend/store"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	repository, err := store.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer repository.Close()
	handler := api.NewServer(service.New(repository, nil), cfg.MaxBodyBytes)
	server := &http.Server{Addr: cfg.Address, Handler: handler, ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()
	log.Printf("knowledge backend %s", cfg.RuntimeSummary())
	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
