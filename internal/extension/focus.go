package extension

import (
	"context"
	"fmt"

	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/focus/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/focus/v1/focusv1connect"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	focusBucketKey = []byte("focus")
	focusKey       = []byte("focus")
)

type focusStore struct {
	DB *bolt.DB
}

func (s *focusStore) save(focus *focusv1.Focus) error {
	data, err := proto.Marshal(focus)
	if err != nil {
		return fmt.Errorf("save focus: %w", err)
	}

	if err := s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(focusBucketKey)
		if err != nil {
			return err
		}
		return b.Put(focusKey, data)
	}); err != nil {
		return fmt.Errorf("save focus: %w", err)
	}
	return nil
}

func (s *focusStore) load() (*focusv1.Focus, error) {
	var data []byte
	if err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(focusBucketKey)
		if b == nil {
			return nil
		}
		data = b.Get(focusKey)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("load focus: %w", err)
	}

	focus := &focusv1.Focus{}
	if err := proto.Unmarshal(data, focus); err != nil {
		return nil, fmt.Errorf("load focus: %w", err)
	}
	return focus, nil
}

type FocusService struct {
	store *focusStore
}

func (s *FocusService) Save(ctx context.Context, req *focusv1.SaveRequest) (*focusv1.SaveResponse, error) {
	if err := s.store.save(req.GetFocus()); err != nil {
		return nil, err
	}
	return &focusv1.SaveResponse{}, nil
}

var _ focusv1connect.FocusServiceHandler = (*FocusService)(nil)

type FocusedProblemLoader struct {
	store *focusStore
}

func (l *FocusedProblemLoader) Load() (*problemv1.Problem, error) {
	focus, err := l.store.load()
	if err != nil {
		return nil, fmt.Errorf("load focused problem: %w", err)
	}

	if errMsg := focus.GetError(); errMsg != "" {
		return nil, fmt.Errorf("load focused problem: %w", &ExtensionError{Msg: errMsg})
	}

	return focus.GetProblem(), nil
}
