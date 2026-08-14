package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/EthanKim8683/cpenv/internal/config"
	extension "github.com/EthanKim8683/cpenv/internal/daemon"
	"github.com/EthanKim8683/cpenv/internal/gen/focus/v1/focusv1connect"
	"github.com/EthanKim8683/cpenv/internal/gen/status/v1/statusv1connect"
	"github.com/EthanKim8683/cpenv/internal/gen/submit/v1/submitv1connect"
	"github.com/adrg/xdg"
	"github.com/rs/cors"
	bolt "go.etcd.io/bbolt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("cpenvd: %v", err)
	}

	dbPath := filepath.Join(xdg.StateHome, "cpenv", "cpenv.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		log.Fatalf("cpenvd: %v", err)
	}

	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatalf("cpenvd: %v", err)
	}

	focusSvc := &extension.FocusService{DB: db}
	statusSvc := &extension.StatusService{DB: db}
	submitSvc := extension.NewSubmitService()

	mux := http.NewServeMux()
	mux.Handle(focusv1connect.NewFocusServiceHandler(focusSvc))
	mux.Handle(statusv1connect.NewStatusServiceHandler(statusSvc))
	mux.Handle(submitv1connect.NewSubmitServiceHandler(submitSvc))

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
	}).Handler(mux)

	srv := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", cfg.Port),
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("cpenvd: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("cpenvd: %v", err)
	}
}
