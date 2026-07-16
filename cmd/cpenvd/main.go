package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	observationv1 "github.com/EthanKim8683/cpenv/gen/observation/v1"
	"github.com/EthanKim8683/cpenv/gen/observation/v1/observationv1connect"
	"github.com/EthanKim8683/cpenv/internal/server"
)

func main() {
	observation := &server.ObservationService{
		OnReportContest: func(_ context.Context, req *observationv1.ReportContestRequest) error {
			fmt.Printf("ReportContest: %+v\n", req)
			return nil
		},
		OnReportProblem: func(_ context.Context, req *observationv1.ReportProblemRequest) error {
			fmt.Printf("ReportProblem: %+v\n", req)
			return nil
		},
		OnFocusTab: func(_ context.Context, req *observationv1.FocusTabRequest) error {
			fmt.Printf("FocusTab: %+v\n", req)
			return nil
		},
	}

	mux := http.NewServeMux()
	mux.Handle(observationv1connect.NewObservationServiceHandler(observation))

	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: mux,
	}
	server.ListenAndServe()
}
