package server_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	submitv1 "github.com/EthanKim8683/cpenv/gen/submit/v1"
	"github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
	"github.com/EthanKim8683/cpenv/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testServer(t *testing.T) *httptest.Server {
	t.Helper()

	svc := server.NewSubmitService()

	mux := http.NewServeMux()
	mux.Handle(submitv1connect.NewSubmitServiceHandler(svc))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func subscribeStream(t *testing.T, client submitv1connect.SubmitServiceClient, tabId string) (*connect.ServerStreamForClient[submitv1.SubscribeResponse], error) {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	stream, err := client.Subscribe(ctx, &submitv1.SubscribeRequest{
		TabId: tabId,
	})
	if err != nil {
		return nil, err
	}

	require.True(t, stream.Receive())
	t.Cleanup(func() {
		stream.Close()
	})
	return stream, nil
}

func submit(t *testing.T, client submitv1connect.SubmitServiceClient, tabId string) error {
	t.Helper()

	_, err := client.Submit(t.Context(), &submitv1.SubmitRequest{
		TabId: tabId,
	})
	return err
}

func receive(t *testing.T, stream *connect.ServerStreamForClient[submitv1.SubscribeResponse]) error {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	ch := make(chan error, 1)
	go func() {
		stream.Receive()
		ch <- stream.Err()
	}()
	select {
	case received := <-ch:
		return received
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestSubmitService(t *testing.T) {
	t.Parallel()

	t.Run("submit forwards to matching subscriber", func(t *testing.T) {
		t.Parallel()

		srv := testServer(t)
		client := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

		stream, err := subscribeStream(t, client, "foo")
		require.NoError(t, err)

		err = submit(t, client, "foo")
		require.NoError(t, err)

		assert.NoError(t, receive(t, stream))
	})

	t.Run("submit does not forward to non-matching subscribers", func(t *testing.T) {
		t.Parallel()

		srv := testServer(t)
		client := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

		fooStream, err := subscribeStream(t, client, "foo")
		require.NoError(t, err)

		barStream, err := subscribeStream(t, client, "bar")
		require.NoError(t, err)

		err = submit(t, client, "foo")
		require.NoError(t, err)

		assert.NoError(t, receive(t, fooStream))
		assert.ErrorIs(t, receive(t, barStream), context.DeadlineExceeded)
	})

	t.Run("submit errors if subscriber not found", func(t *testing.T) {
		t.Parallel()

		srv := testServer(t)
		client := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

		stream, err := subscribeStream(t, client, "foo")
		require.NoError(t, err)

		err = submit(t, client, "bar")
		connectErr, ok := errors.AsType[*connect.Error](err)
		require.True(t, ok)
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())

		assert.ErrorIs(t, receive(t, stream), context.DeadlineExceeded)
	})

	t.Run("subscribe replaces existing subscriber", func(t *testing.T) {
		t.Parallel()

		srv := testServer(t)
		client := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

		oldStream, err := subscribeStream(t, client, "foo")
		require.NoError(t, err)

		newStream, err := subscribeStream(t, client, "foo")
		require.NoError(t, err)

		err = submit(t, client, "foo")
		require.NoError(t, err)

		assert.NoError(t, receive(t, newStream))

		err = receive(t, oldStream)
		connectErr, ok := errors.AsType[*connect.Error](err)
		require.True(t, ok)
		assert.Equal(t, connect.CodeCanceled, connectErr.Code())
	})
}
