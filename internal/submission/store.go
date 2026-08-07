package submission

import (
	"fmt"

	submissionv1 "github.com/EthanKim8683/cpenv/internal/gen/submission/v1"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	bucketKeyPrimary   = []byte("submissions")
	bucketKeyByProblem = []byte("submissions_by_problem")
)

type store struct {
	db *bolt.DB
}

func key(ms int64, id string) []byte {
	k := make([]byte, 6+len(id))
	k[0] = byte(ms >> 40)
	k[1] = byte(ms >> 32)
	k[2] = byte(ms >> 24)
	k[3] = byte(ms >> 16)
	k[4] = byte(ms >> 8)
	k[5] = byte(ms)
	copy(k[6:], id)
	return k
}

func (s *store) save(subs []*submissionv1.Submission) error {
	if err := s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketKeyPrimary)
		if err != nil {
			return err
		}
		pbb, err := tx.CreateBucketIfNotExists(bucketKeyByProblem)
		if err != nil {
			return err
		}

		for _, sub := range subs {
			data, err := proto.Marshal(sub)
			if err != nil {
				return err
			}
			k := key(sub.GetTimestampMs(), sub.GetProblemId())
			if err := b.Put(k, data); err != nil {
				return err
			}

			pb, err := pbb.CreateBucketIfNotExists([]byte(sub.GetProblemId()))
			if err != nil {
				return err
			}
			if err := pb.Put(k, nil); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("save submissions: %w", err)
	}
	return nil
}

func (s *store) tail(limit int) ([]*submissionv1.Submission, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("tail submissions: limit must be positive: %d", limit)
	}

	subs := []*submissionv1.Submission{}
	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketKeyPrimary)
		if b == nil {
			return nil
		}

		subs = make([]*submissionv1.Submission, limit)
		c := b.Cursor()
		k, v := c.Last()
		for i := range limit {
			if k == nil {
				subs = subs[limit-i:]
				return nil
			}

			sub := &submissionv1.Submission{}
			if err := proto.Unmarshal(v, sub); err != nil {
				return err
			}
			subs[limit-1-i] = sub

			k, v = c.Prev()
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("tail submissions: %w", err)
	}
	return subs, nil
}

func (s *store) tailProblem(problemID string, limit int) ([]*submissionv1.Submission, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("tail problem %q submissions: limit must be positive: %d", problemID, limit)
	}

	subs := []*submissionv1.Submission{}
	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketKeyPrimary)
		if b == nil {
			return nil
		}

		pbb := tx.Bucket(bucketKeyByProblem)
		if pbb == nil {
			return nil
		}
		pb := pbb.Bucket([]byte(problemID))
		if pb == nil {
			return nil
		}

		subs = make([]*submissionv1.Submission, limit)
		c := pb.Cursor()
		k, _ := c.Last()
		for i := range limit {
			if k == nil {
				subs = subs[limit-i:]
				return nil
			}

			sub := &submissionv1.Submission{}
			if err := proto.Unmarshal(b.Get(k), sub); err != nil {
				return err
			}
			subs[limit-1-i] = sub

			k, _ = c.Prev()
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("tail problem %q submissions: %w", problemID, err)
	}
	return subs, nil
}
