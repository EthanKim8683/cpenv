package extension

import (
	"context"
	"errors"
	"fmt"

	statusv1 "github.com/EthanKim8683/cpenv/internal/gen/status/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/status/v1/statusv1connect"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	submissionsBucketKey          = []byte("submissions")
	submissionsByProblemBucketKey = []byte("submissions_by_problem")
)

type statusStore struct {
	DB *bolt.DB
}

func submissionKey(sub *statusv1.Submission) []byte {
	ts := sub.GetTimestampMs()
	pID := sub.GetProblemId()

	buf := make([]byte, 6+len(pID))
	buf[0] = byte(ts >> 40)
	buf[1] = byte(ts >> 32)
	buf[2] = byte(ts >> 24)
	buf[3] = byte(ts >> 16)
	buf[4] = byte(ts >> 8)
	buf[5] = byte(ts)
	copy(buf[6:], pID)
	return buf
}

func saveSubmission(b *bolt.Bucket, pbb *bolt.Bucket, sub *statusv1.Submission) error {
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

	return pb.Put(key, data)
}

func (s *statusStore) save(subs []*statusv1.Submission) error {
	if err := s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(submissionsBucketKey)
		if err != nil {
			return err
		}

		pbb, err := tx.CreateBucketIfNotExists(submissionsByProblemBucketKey)
		if err != nil {
			return err
		}

		var errs error
		for _, sub := range subs {
			errs = errors.Join(errs, saveSubmission(b, pbb, sub))
		}
		return errs
	}); err != nil {
		return fmt.Errorf("save submissions: %w", err)
	}
	return nil
}

func (s *statusStore) tail(limit int) ([]*statusv1.Submission, error) {
	subs := []*statusv1.Submission{}
	if err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(submissionsBucketKey)
		if b == nil {
			return nil
		}

		subs = make([]*statusv1.Submission, limit)
		var errs error
		c := b.Cursor()
		k, v := c.Last()
		for i := range limit {
			if k == nil {
				subs = subs[i+1:]
				break
			}

			if err := proto.Unmarshal(v, subs[i]); err != nil {
				errs = errors.Join(errs)
			}

			k, v = c.Prev()
		}
		return errs
	}); err != nil {
		return nil, fmt.Errorf("tail submissions: %w", err)
	}
	return subs, nil
}

func (s *statusStore) tailProblem(problemID string, limit int) ([]*statusv1.Submission, error) {
	subs := []*statusv1.Submission{}
	if err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(submissionsBucketKey)
		if b == nil {
			return nil
		}

		pbb := tx.Bucket(submissionsByProblemBucketKey)
		if pbb == nil {
			return nil
		}
		pb := pbb.Bucket([]byte(problemID))
		if pb == nil {
			return nil
		}

		subs = make([]*statusv1.Submission, limit)
		var errs error
		c := pb.Cursor()
		k, _ := c.Last()
		for i := range limit {
			if k == nil {
				subs = subs[i+1:]
				break
			}

			if err := proto.Unmarshal(b.Get(k), subs[i]); err != nil {
				errs = errors.Join(errs)
			}

			k, _ = c.Prev()
		}
		return errs
	}); err != nil {
		return nil, fmt.Errorf("tail submissions for %q: %w", problemID, err)
	}
	return subs, nil
}

type StatusService struct {
	store *statusStore
}

func (s *StatusService) Save(ctx context.Context, req *statusv1.SaveRequest) (*statusv1.SaveResponse, error) {
	if err := s.store.save(req.GetSubmissions()); err != nil {
		return nil, err
	}
	return &statusv1.SaveResponse{}, nil
}

var _ statusv1connect.StatusServiceHandler = (*StatusService)(nil)

type SubmissionsTailer struct {
	store *statusStore
}

func (t *SubmissionsTailer) Tail(limit int) ([]*statusv1.Submission, error) {
	return t.store.tail(limit)
}

func (t *SubmissionsTailer) TailProblem(problemID string, limit int) ([]*statusv1.Submission, error) {
	return t.store.tailProblem(problemID, limit)
}
