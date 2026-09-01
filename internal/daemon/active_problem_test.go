package daemon

import (
	"path/filepath"
	"testing"

	activeproblemv1 "github.com/EthanKim8683/cpenv/internal/gen/active_problem/v1"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

func TestActiveProblemService(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		activeProblem := &activeproblemv1.ActiveProblem{
			Problem: &problemv1.Problem{Id: "id"},
			Error:   new("error"),
		}

		db, err := bolt.Open(filepath.Join(t.TempDir(), "db.db"), 0600, nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})

		svc := &ActiveProblemService{DB: db}

		require.NoError(t, svc.save(activeProblem))

		gotActiveProblem, err := svc.load()
		require.NoError(t, err)
		assert.True(t, proto.Equal(activeProblem, gotActiveProblem))
	})
}
