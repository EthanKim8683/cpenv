package daemon

import (
	"path/filepath"
	"testing"

	statusv1 "github.com/EthanKim8683/cpenv/internal/gen/status/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestStatus(t *testing.T) {
	t.Parallel()

	db, err := bolt.Open(filepath.Join(t.TempDir(), "db.db"), 0600, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	svc := &StatusService{DB: db}

	t.Run("round trip", func(t *testing.T) {
		subs := []*statusv1.Submission{
			{TimestampMs: 1, ProblemId: "1"},
			{TimestampMs: 1, ProblemId: "2", Status: statusv1.Status_STATUS_PENDING},
			{TimestampMs: 2, ProblemId: "1"},
		}

		err = svc.save(subs)
		require.NoError(t, err)

		subs, err := svc.tail(defaultTailLimit, nil)
		require.NoError(t, err)
		assert.Empty(t, cmp.Diff(subs, subs, protocmp.Transform()))
	})

	t.Run("tail with limit", func(t *testing.T) {
		limit := 2
		subs := []*statusv1.Submission{
			{TimestampMs: 1, ProblemId: "2", Status: statusv1.Status_STATUS_PENDING},
			{TimestampMs: 2, ProblemId: "1"},
		}

		subs, err := svc.tail(limit, nil)
		require.NoError(t, err)
		assert.Empty(t, cmp.Diff(subs, subs, protocmp.Transform()))
	})

	t.Run("tail with problem ID", func(t *testing.T) {
		problemID := "1"
		subs := []*statusv1.Submission{
			{TimestampMs: 1, ProblemId: "1"},
			{TimestampMs: 2, ProblemId: "1"},
		}

		subs, err := svc.tail(defaultTailLimit, &problemID)
		require.NoError(t, err)
		assert.Empty(t, cmp.Diff(subs, subs, protocmp.Transform()))
	})

	t.Run("overwrite", func(t *testing.T) {
		saveSubs := []*statusv1.Submission{
			{TimestampMs: 1, ProblemId: "2", Status: statusv1.Status_STATUS_ACCEPTED},
		}
		loadSubs := []*statusv1.Submission{
			{TimestampMs: 1, ProblemId: "1"},
			{TimestampMs: 1, ProblemId: "2", Status: statusv1.Status_STATUS_ACCEPTED},
			{TimestampMs: 2, ProblemId: "1"},
		}

		err := svc.save(saveSubs)
		require.NoError(t, err)

		subs, err := svc.tail(defaultTailLimit, nil)
		require.NoError(t, err)
		assert.Empty(t, cmp.Diff(loadSubs, subs, protocmp.Transform()))
	})
}
