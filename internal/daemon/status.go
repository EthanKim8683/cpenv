package daemon

import (
	"context"

	statusv1 "github.com/EthanKim8683/cpenv/internal/gen/status/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/status/v1/statusv1connect"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

const (
	defaultTailLimit = 10
	maxTailLimit     = 100
)

var (
	submissionsBucketKey          = []byte("submissions")
	submissionsByProblemBucketKey = []byte("submissions_by_problem")
)

type StatusService struct {
	DB *bolt.DB
}

func submissionKey(sub *statusv1.Submission) []byte {
	ts := sub.GetTimestampMs()
	id := sub.GetProblemId()
	key := make([]byte, 6+len(id))
	key[0] = byte(ts >> 40)
	key[1] = byte(ts >> 32)
	key[2] = byte(ts >> 24)
	key[3] = byte(ts >> 16)
	key[4] = byte(ts >> 8)
	key[5] = byte(ts)
	copy(key[6:], id)
	return key
}

func (s *StatusService) Save(ctx context.Context, req *statusv1.SaveRequest) (*statusv1.SaveResponse, error) {
	if err := s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(submissionsBucketKey)
		if err != nil {
			return err
		}

		pbb, err := tx.CreateBucketIfNotExists(submissionsByProblemBucketKey)
		if err != nil {
			return err
		}

		for _, sub := range req.GetSubmissions() {
			data, err := proto.Marshal(sub)
			if err != nil {
				return err
			}

			key := submissionKey(sub)
			if err := b.Put(key, data); err != nil {
				return err
			}

			pb, err := pbb.CreateBucketIfNotExists([]byte(sub.GetProblemId()))
			if err != nil {
				return err
			}

			if err := pb.Put(key, nil); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &statusv1.SaveResponse{}, nil
}

func (s *StatusService) Tail(ctx context.Context, req *statusv1.TailRequest) (*statusv1.TailResponse, error) {
	limit := defaultTailLimit
	if req.Limit != nil {
		limit = min(int(req.GetLimit()), maxTailLimit)
	}

	subs := []*statusv1.Submission{}
	if err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(submissionsBucketKey)
		if b == nil {
			return nil
		}

		if req.ProblemId == nil {
			subs = make([]*statusv1.Submission, limit)
			c := b.Cursor()
			k, v := c.Last()
			for i := limit - 1; i >= 0; i-- {
				if k == nil {
					subs = subs[i+1:]
					break
				}
				subs[i] = &statusv1.Submission{}
				if err := proto.Unmarshal(v, subs[i]); err != nil {
					return err
				}
				k, v = c.Prev()
			}
		} else {
			pbb := tx.Bucket(submissionsByProblemBucketKey)
			if pbb == nil {
				return nil
			}

			pb := pbb.Bucket([]byte(req.GetProblemId()))
			if pb == nil {
				return nil
			}

			subs = make([]*statusv1.Submission, limit)
			c := pb.Cursor()
			k, _ := c.Last()
			for i := limit - 1; i >= 0; i-- {
				if k == nil {
					subs = subs[i+1:]
					break
				}
				subs[i] = &statusv1.Submission{}
				if err := proto.Unmarshal(b.Get(k), subs[i]); err != nil {
					return err
				}
				k, _ = c.Prev()
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &statusv1.TailResponse{Submissions: subs}, nil
}

var _ statusv1connect.StatusServiceHandler = (*StatusService)(nil)
