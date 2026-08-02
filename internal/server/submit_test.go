package server_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	submitv1 "github.com/EthanKim8683/cpenv/gen/submit/v1"
	submitv1connect "github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
	"github.com/EthanKim8683/cpenv/internal/server"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type subscribeFunc func(cb func(req *submitv1.SubscribeResponse) *submitv1.CallbackRequest)
type submitFunc func(fileName string, content []byte) error

func TestSubmitService(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		extension func(t *testing.T, subscribe subscribeFunc)
		cli       func(t *testing.T, submit submitFunc)
	}{
		"successful submission": {
			extension: func(t *testing.T, subscribe subscribeFunc) {
				subscribe(func(req *submitv1.SubscribeResponse) *submitv1.CallbackRequest {
					assert.Equal(t, "filename", req.FileName)
					assert.Equal(t, []byte("content"), req.Content)
					return &submitv1.CallbackRequest{
						CallbackId: req.CallbackId,
					}
				})
			},
			cli: func(t *testing.T, submit submitFunc) {
				assert.NoError(t, submit("filename", []byte("content")))
			},
		},
		"no subscribers": {
			extension: func(t *testing.T, subscribe subscribeFunc) {},
			cli: func(t *testing.T, submit submitFunc) {
				assert.ErrorContains(t, submit("", nil), "no subscribers")
			},
		},
		"extension error": {
			extension: func(t *testing.T, subscribe subscribeFunc) {
				subscribe(func(req *submitv1.SubscribeResponse) *submitv1.CallbackRequest {
					return &submitv1.CallbackRequest{
						CallbackId: req.CallbackId,
						Error:      "error message",
					}
				})
			},
			cli: func(t *testing.T, submit submitFunc) {
				assert.ErrorContains(t, submit("", nil), "extension: error message")
			},
		},
		"stress test": {
			extension: func(t *testing.T, subscribe subscribeFunc) {
				for range 10 {
					subscribe(func(req *submitv1.SubscribeResponse) *submitv1.CallbackRequest {
						return &submitv1.CallbackRequest{
							CallbackId: req.CallbackId,
							Error:      req.FileName,
						}
					})
				}
			},
			cli: func(t *testing.T, submit submitFunc) {
				var wg sync.WaitGroup
				for range 20 {
					wg.Go(func() {
						id := uuid.New().String()
						assert.ErrorContains(t, submit(id, nil), id)
					})
				}
				wg.Wait()
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := server.NewSubmitService()

			mux := http.NewServeMux()
			mux.Handle(submitv1connect.NewSubmitServiceHandler(svc))

			srv := httptest.NewServer(mux)
			t.Cleanup(srv.Close)

			ctx, cancel := context.WithCancel(t.Context())
			t.Cleanup(cancel)

			problemID := t.Name()

			var wg sync.WaitGroup
			subscribe := func(cb func(req *submitv1.SubscribeResponse) *submitv1.CallbackRequest) {
				wg.Go(func() {
					ext := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

					stream, err := ext.Subscribe(ctx, &submitv1.SubscribeRequest{
						ProblemId: problemID,
					})
					if errors.Is(err, context.Canceled) {
						return
					}
					require.NoError(t, err)

					t.Cleanup(func() {
						err := stream.Close()
						if errors.Is(err, context.Canceled) {
							return
						}
						require.NoError(t, err)
					})

					for {
						if !stream.Receive() {
							return
						}

						_, err = ext.Callback(ctx, cb(stream.Msg()))
						if errors.Is(err, context.Canceled) {
							return
						}
						require.NoError(t, err)
					}
				})
			}

			submit := func(fileName string, content []byte) error {
				tick := 100 * time.Millisecond
				timeout := 1 * time.Second

				cli := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

				deadline := time.Now().Add(timeout)
				var err error
				for {
					if time.Now().After(deadline) {
						break
					}

					if _, err = cli.Submit(ctx, &submitv1.SubmitRequest{
						ProblemId: problemID,
						FileName:  fileName,
						Content:   content,
					}); err == nil || !strings.Contains(err.Error(), "no subscribers") {
						break
					}

					time.Sleep(tick)
				}
				return err
			}

			if test.extension != nil {
				test.extension(t, subscribe)
			}
			if test.cli != nil {
				test.cli(t, submit)
			}
			cancel()
		})
	}
}
