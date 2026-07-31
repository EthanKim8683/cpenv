package server_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	submitv1 "github.com/EthanKim8683/cpenv/gen/submit/v1"
	submitv1connect "github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
	"github.com/EthanKim8683/cpenv/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func serveSubscriber(
	t *testing.T,
	ctx context.Context,
	srv *httptest.Server,
	req *submitv1.SubscribeRequest,
	cb func(req *submitv1.SubscribeResponse) *submitv1.CallbackRequest,
) {
	t.Helper()

	ext := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

	stream, err := ext.Subscribe(ctx, req)
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
}

func eventuallySubmit(
	t *testing.T,
	cli submitv1connect.SubmitServiceClient,
	req *submitv1.SubmitRequest,
) error {
	t.Helper()

	var err error
	require.Eventually(t, func() bool {
		_, err = cli.Submit(t.Context(), req)
		return !(err != nil && strings.Contains(err.Error(), "no subscribers"))
	}, 2*time.Second, 100*time.Millisecond)
	return err
}

func TestSubmitService(t *testing.T) {
	t.Parallel()

	svc := server.NewSubmitService()

	mux := http.NewServeMux()
	mux.Handle(submitv1connect.NewSubmitServiceHandler(svc))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	path := filepath.Join(t.TempDir(), "path")
	data := []byte("data")
	require.NoError(t, os.WriteFile(path, data, 0644))

	t.Run("successful submit", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(cancel)

		problemID := t.Name()

		var wg sync.WaitGroup
		wg.Go(func() {
			serveSubscriber(
				t,
				ctx,
				srv,
				&submitv1.SubscribeRequest{
					ProblemId: problemID,
				},
				func(req *submitv1.SubscribeResponse) *submitv1.CallbackRequest {
					assert.Equal(t, path, req.Path)
					assert.Equal(t, data, req.Data)
					return &submitv1.CallbackRequest{
						CallbackId: req.CallbackId,
					}
				},
			)
		})
		wg.Go(func() {
			cli := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

			assert.NoError(t, eventuallySubmit(t, cli, &submitv1.SubmitRequest{
				ProblemId: problemID,
				Path:      path,
			}))

			cancel()
		})
		wg.Wait()
	})

	t.Run("no subscribers", func(t *testing.T) {
		t.Parallel()

		problemID := t.Name()

		cli := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

		_, err := cli.Submit(t.Context(), &submitv1.SubmitRequest{
			ProblemId: problemID,
			Path:      path,
		})
		assert.ErrorContains(t, err, "no subscribers")
	})

	t.Run("extension error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(cancel)

		problemID := t.Name()
		errMsg := t.Name()

		var wg sync.WaitGroup
		wg.Go(func() {
			serveSubscriber(
				t,
				ctx,
				srv,
				&submitv1.SubscribeRequest{
					ProblemId: problemID,
				},
				func(req *submitv1.SubscribeResponse) *submitv1.CallbackRequest {
					return &submitv1.CallbackRequest{
						CallbackId: req.CallbackId,
						Error:      errMsg,
					}
				},
			)
		})
		wg.Go(func() {
			cli := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

			err := eventuallySubmit(t, cli, &submitv1.SubmitRequest{
				ProblemId: problemID,
				Path:      path,
			})
			assert.ErrorContains(t, err, fmt.Sprintf("extension: %s", errMsg))

			cancel()
		})
		wg.Wait()
	})

	t.Run("multiple subscribers and concurrent submits", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(cancel)

		problemID := t.Name()
		subscribers := 10
		submits := 20

		var count atomic.Int32
		var wg sync.WaitGroup
		for range subscribers {
			wg.Go(func() {
				serveSubscriber(
					t,
					ctx,
					srv,
					&submitv1.SubscribeRequest{
						ProblemId: problemID,
					},
					func(req *submitv1.SubscribeResponse) *submitv1.CallbackRequest {
						count.Add(1)
						return &submitv1.CallbackRequest{
							CallbackId: req.CallbackId,
						}
					},
				)
			})
		}
		wg.Go(func() {
			cli := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

			var submitWG sync.WaitGroup
			for range submits {
				submitWG.Go(func() {
					assert.NoError(t, eventuallySubmit(t, cli, &submitv1.SubmitRequest{
						ProblemId: problemID,
						Path:      path,
					}))
				})
			}
			submitWG.Wait()

			cancel()
		})
		wg.Wait()
		assert.Equal(t, submits, int(count.Load()))
	})
}
