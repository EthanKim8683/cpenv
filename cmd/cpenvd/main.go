package main

import (
	"log"
	"net/http"
	"os"

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
		log.Fatalf("cpenv: config: %v", err)
	}

	stateStore := state.NewStore(cfg.StatePath)

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

	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: handler,
	}
	server.ListenAndServe()
}
