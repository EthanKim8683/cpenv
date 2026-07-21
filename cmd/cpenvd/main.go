package main

import (
	"net/http"
	"os"

	"github.com/EthanKim8683/cpenv/gen/focus/v1/focusv1connect"
	"github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
	"github.com/EthanKim8683/cpenv/internal/server"
	"github.com/rs/cors"
)

func main() {
	focusSvc := &server.FocusService{}
	submitSvc := &server.SubmitService{}

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
