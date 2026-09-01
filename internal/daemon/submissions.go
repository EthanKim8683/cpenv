package daemon

import (
	"context"

	submissionsv1 "github.com/EthanKim8683/cpenv/internal/gen/submissions/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/submissions/v1/submissionsv1connect"
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

type SubmissionsService struct {
	DB *bolt.DB
}

func submissionKey(sub *submissionsv1.Submission) []byte {
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

func (s *SubmissionsService) save(subs []*submissionsv1.Submission) error {
	if err := s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(submissionsBucketKey)
		if err != nil {
			return err
		}

		pbb, err := tx.CreateBucketIfNotExists(submissionsByProblemBucketKey)
		if err != nil {
			return err
		}

		for _, sub := range subs {
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
		return err
	}
	return nil
}

func (s *SubmissionsService) tail(limit int, problemID *string) ([]*submissionsv1.Submission, error) {
	subs := []*submissionsv1.Submission{}
	if err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(submissionsBucketKey)
		if b == nil {
			return nil
		}

		var c *bolt.Cursor
		if problemID == nil {
			c = b.Cursor()
		} else {
			pbb := tx.Bucket(submissionsByProblemBucketKey)
			if pbb == nil {
				return nil
			}
			pb := pbb.Bucket([]byte(*problemID))
			if pb == nil {
				return nil
			}
			c = pb.Cursor()
		}

		subs = make([]*submissionsv1.Submission, limit)
		k, v := c.Last()
		for i := limit - 1; i >= 0; i-- {
			if k == nil {
				subs = subs[i+1:]
				break
			}
			if len(v) == 0 {
				v = b.Get(k)
			}
			subs[i] = &submissionsv1.Submission{}
			if err := proto.Unmarshal(v, subs[i]); err != nil {
				return err
			}
			k, v = c.Prev()
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return subs, nil
}

func (s *SubmissionsService) Save(_ context.Context, req *submissionsv1.SaveRequest) (*submissionsv1.SaveResponse, error) {
	if err := s.save(req.GetSubmissions()); err != nil {
		return nil, err
	}
	return &submissionsv1.SaveResponse{}, nil
}

func (s *SubmissionsService) Tail(_ context.Context, req *submissionsv1.TailRequest) (*submissionsv1.TailResponse, error) {
	limit := defaultTailLimit
	if req.Limit != nil {
		limit = min(int(req.GetLimit()), maxTailLimit)
	}

	subs, err := s.tail(limit, req.ProblemId)
	if err != nil {
		return nil, err
	}
	return &submissionsv1.TailResponse{Submissions: subs}, nil
}

var _ submissionsv1connect.SubmissionsServiceHandler = (*SubmissionsService)(nil)
