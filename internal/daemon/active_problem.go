package daemon

import (
	"context"

	activeproblemv1 "github.com/EthanKim8683/cpenv/internal/gen/active_problem/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/active_problem/v1/active_problemv1connect"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	activeProblemBucketKey = []byte("active_problem")
	activeProblemKey       = []byte("active_problem")
)

type ActiveProblemService struct {
	DB *bolt.DB
}

func (s *ActiveProblemService) save(activeProblem *activeproblemv1.ActiveProblem) error {
	data, err := proto.Marshal(activeProblem)
	if err != nil {
		return err
	}

	if err := s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(activeProblemBucketKey)
		if err != nil {
			return err
		}
		return b.Put(activeProblemKey, data)
	}); err != nil {
		return err
	}
	return nil
}

func (s *ActiveProblemService) load() (*activeproblemv1.ActiveProblem, error) {
	var data []byte
	if err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(activeProblemBucketKey)
		if b == nil {
			return nil
		}
		data = b.Get(activeProblemKey)
		return nil
	}); err != nil {
		return nil, err
	}

	activeProblem := &activeproblemv1.ActiveProblem{}
	if err := proto.Unmarshal(data, activeProblem); err != nil {
		return nil, err
	}
	return activeProblem, nil
}

func (s *ActiveProblemService) Save(_ context.Context, req *activeproblemv1.SaveRequest) (*activeproblemv1.SaveResponse, error) {
	if err := s.save(req.GetActiveProblem()); err != nil {
		return nil, err
	}
	return &activeproblemv1.SaveResponse{}, nil
}

func (s *ActiveProblemService) Load(_ context.Context, req *activeproblemv1.LoadRequest) (*activeproblemv1.LoadResponse, error) {
	activeProblem, err := s.load()
	if err != nil {
		return nil, err
	}
	return &activeproblemv1.LoadResponse{ActiveProblem: activeProblem}, nil
}

var _ active_problemv1connect.ActiveProblemServiceHandler = (*ActiveProblemService)(nil)
