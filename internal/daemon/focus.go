package daemon

import (
	"context"

	"connectrpc.com/connect"
	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/focus/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/focus/v1/focusv1connect"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	focusBucketKey = []byte("focus")
	focusKey       = []byte("focus")
)

type FocusService struct {
	DB *bolt.DB
}

func (s *FocusService) save(focus *focusv1.Focus) error {
	data, err := proto.Marshal(focus)
	if err != nil {
		return err
	}

	if err := s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(focusBucketKey)
		if err != nil {
			return err
		}
		return b.Put(focusKey, data)
	}); err != nil {
		return err
	}
	return nil
}

func (s *FocusService) load() (*focusv1.Focus, error) {
	var data []byte
	if err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(focusBucketKey)
		if b == nil {
			return nil
		}
		data = b.Get(focusKey)
		return nil
	}); err != nil {
		return nil, err
	}

	focus := &focusv1.Focus{}
	if err := proto.Unmarshal(data, focus); err != nil {
		return nil, err
	}
	return focus, nil
}

func (s *FocusService) Save(_ context.Context, req *focusv1.SaveRequest) (*focusv1.SaveResponse, error) {
	if err := s.save(req.GetFocus()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return &focusv1.SaveResponse{}, nil
}

func (s *FocusService) Load(_ context.Context, req *focusv1.LoadRequest) (*focusv1.LoadResponse, error) {
	focus, err := s.load()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return &focusv1.LoadResponse{Focus: focus}, nil
}

var _ focusv1connect.FocusServiceHandler = (*FocusService)(nil)
