package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/EthanKim8683/cpenv/gen/focus/v1/focusv1connect"
	"github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/EthanKim8683/cpenv/internal/server"
	"github.com/EthanKim8683/cpenv/internal/state"
	"github.com/rs/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("cpenvd: %v", err)
	}

	stateStore := state.NewFileStore(cfg.StatePath)

	focusSvc := &server.FocusService{
		StateStore: stateStore,
	}
	submitSvc := server.NewSubmitService()

	mux := http.NewServeMux()
	mux.Handle(focusv1connect.NewFocusServiceHandler(focusSvc))
	mux.Handle(submitv1connect.NewSubmitServiceHandler(submitSvc))

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
	}).Handler(mux)

	srv := &http.Server{
		Addr:    "localhost:" + cfg.Port,
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

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("cpenvd: %v", err)
	}
}
