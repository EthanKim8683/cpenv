package submission

import (
	"path/filepath"
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	submissionv1 "github.com/EthanKim8683/cpenv/internal/gen/submission/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func equalProto[T proto.Message](t *testing.T, expected []T, actual []T) bool {
	t.Helper()
	if len(actual) != len(expected) {
		return false
	}
	for i := range expected {
		if !proto.Equal(expected[i], actual[i]) {
			return false
		}
	}
	return true
}

func elementsMatchProto[T proto.Message](t *testing.T, expected []T, actual []T) bool {
	t.Helper()
	if len(actual) != len(expected) {
		return false
	}
	used := make([]bool, len(actual))
	for _, e := range expected {
		found := false
		for i, a := range actual {
			if !used[i] && proto.Equal(e, a) {
				found = true
				used[i] = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestDBStore(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "db.db")
		subs := []*submissionv1.Submission{
			{TimestampMs: 0, ProblemId: "1"},
			{TimestampMs: 1, ProblemId: "1"},
		}

		db, err := bolt.Open(path, 0600, nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})
		s := &DBStore{DB: db}
		require.NoError(t, s.save(subs))
		gotSubs, err := s.Tail(2)
		assert.NoError(t, err)
		assert.True(t, equalProto(t, subs, gotSubs))
	})

	t.Run("tail for problem", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "db.db")
		subs1 := []*submissionv1.Submission{
			{TimestampMs: 0, ProblemId: "1"},
			{TimestampMs: 1, ProblemId: "1"},
		}
		subs2 := []*submissionv1.Submission{
			{TimestampMs: 0, ProblemId: "2"},
		}

		db, err := bolt.Open(path, 0600, nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})
		s := &DBStore{DB: db}
		require.NoError(t, s.save(slices.Concat(subs1, subs2)))
		gotSubs, err := s.TailProblem("1", 2)
		assert.NoError(t, err)
		assert.True(t, equalProto(t, subs1, gotSubs))
	})

	t.Run("fewer than limit", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "db.db")
		subs := []*submissionv1.Submission{
			{TimestampMs: 0, ProblemId: "1"},
		}

		db, err := bolt.Open(path, 0600, nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})
		s := &DBStore{DB: db}
		require.NoError(t, s.save(subs))
		require.NoError(t, s.save(subs))
		gotSubs, err := s.Tail(100)
		assert.NoError(t, err)
		assert.True(t, equalProto(t, subs, gotSubs))
	})

	t.Run("matching timestamps", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "db.db")
		subs := []*submissionv1.Submission{
			{TimestampMs: 0, ProblemId: "1"},
			{TimestampMs: 0, ProblemId: "2"},
		}

		db, err := bolt.Open(path, 0600, nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})
		s := &DBStore{DB: db}
		require.NoError(t, s.save(subs))
		gotSubs, err := s.Tail(2)
		assert.NoError(t, err)
		assert.True(t, elementsMatchProto(t, subs, gotSubs))
	})

	t.Run("idempotent", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "db.db")
		subs := []*submissionv1.Submission{
			{TimestampMs: 0, ProblemId: "1"},
			{TimestampMs: 1, ProblemId: "1"},
		}

		db, err := bolt.Open(path, 0600, nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})
		s := &DBStore{DB: db}
		require.NoError(t, s.save(subs))
		require.NoError(t, s.save(subs))
		gotSubs, err := s.Tail(4)
		assert.NoError(t, err)
		assert.True(t, equalProto(t, subs, gotSubs))
	})
}
