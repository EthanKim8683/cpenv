package daemon

import (
	"path/filepath"
	"testing"

	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

func TestFocusService(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		focus := &focusv1.Focus{
			Problem: &problemv1.Problem{Id: "id"},
			Error:   new("error"),
		}

		db, err := bolt.Open(filepath.Join(t.TempDir(), "db.db"), 0600, nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})

		svc := &FocusService{DB: db}

		require.NoError(t, svc.save(focus))

		gotFocus, err := svc.load()
		require.NoError(t, err)
		assert.True(t, proto.Equal(focus, gotFocus))
	})
}
