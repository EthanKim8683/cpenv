package daemon

import (
	"context"

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

func (s *FocusService) Save(_ context.Context, req *focusv1.SaveRequest) (*focusv1.SaveResponse, error) {
	data, err := proto.Marshal(req.GetFocus())
	if err != nil {
		return nil, err
	}

	if err := s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(focusBucketKey)
		if err != nil {
			return err
		}
		return b.Put(focusKey, data)
	}); err != nil {
		return nil, err
	}
	return &focusv1.SaveResponse{}, nil
}

func (s *FocusService) Load(ctx context.Context, req *focusv1.LoadRequest) (*focusv1.LoadResponse, error) {
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
	return &focusv1.LoadResponse{Focus: focus}, nil
}

var _ focusv1connect.FocusServiceHandler = (*FocusService)(nil)
