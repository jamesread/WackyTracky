package migrate

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wacky-tracky/wacky-tracky-server/internal/iamstore"
	"github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

func TestAssertRequiredMissing(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := iamstore.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	err = AssertRequired(context.Background(), st, runtimeconfig.RequiredMigration, dbPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), dbPath)
	require.Contains(t, err.Error(), "migrations table could not be queried")
}
