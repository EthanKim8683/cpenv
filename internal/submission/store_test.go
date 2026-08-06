package submission

import (
	"path/filepath"
	"slices"
	"testing"

	submissionv1 "github.com/EthanKim8683/cpenv/internal/gen/submission/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

func assertProtosEqual(t *testing.T, expected, actual []*submissionv1.Submission) {
	t.Helper()
	require.Equal(t, len(expected), len(actual), "length mismatch")
	for i := range expected {
		assert.True(t, proto.Equal(expected[i], actual[i]), "mismatch at index %d: expected %+v, got %+v", i, expected[i], actual[i])
	}
}

func TestStore(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "db.db")
		subs := []*submissionv1.Submission{
			{TimestampMs: 0, ProblemId: "1"},
			{TimestampMs: 1, ProblemId: "1"},
			{TimestampMs: 3, ProblemId: "1"},
		}

		db, err := bolt.Open(path, 0600, nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})
		s := &store{db: db}
		require.NoError(t, s.save(subs))
		gotSubs, err := s.tail(2)
		assert.NoError(t, err)
		assertProtosEqual(t, subs[1:], gotSubs)
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
		s := &store{db: db}
		require.NoError(t, s.save(slices.Concat(subs1, subs2)))
		gotSubs, err := s.tailProblem("1", 2)
		assert.NoError(t, err)
		assertProtosEqual(t, subs1, gotSubs)
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
		s := &store{db: db}
		require.NoError(t, s.save(subs))
		require.NoError(t, s.save(subs))
		gotSubs, err := s.tail(100)
		assert.NoError(t, err)
		assertProtosEqual(t, subs, gotSubs)
	})

	t.Run("equal timestamps", func(t *testing.T) {
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
		s := &store{db: db}
		require.NoError(t, s.save(subs))
		gotSubs, err := s.tail(2)
		assert.NoError(t, err)
		require.Equal(t, len(subs), len(gotSubs))
		// Check that both elements exist regardless of order
		match1 := (proto.Equal(subs[0], gotSubs[0]) && proto.Equal(subs[1], gotSubs[1]))
		match2 := (proto.Equal(subs[0], gotSubs[1]) && proto.Equal(subs[1], gotSubs[0]))
		assert.True(t, match1 || match2)
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
		s := &store{db: db}
		require.NoError(t, s.save(subs))
		require.NoError(t, s.save(subs))
		gotSubs, err := s.tail(4)
		assert.NoError(t, err)
		assertProtosEqual(t, subs, gotSubs)
	})
}
