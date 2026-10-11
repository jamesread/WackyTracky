package migrate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wacky-tracky/wacky-tracky-server/internal/iamstore"
	"github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

func TestRunSQLiteUp(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	repo := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	t.Chdir(repo)

	migrationsDir, err := FindSQLiteMigrationsDir()
	require.NoError(t, err)

	dbPath := filepath.Join(t.TempDir(), "migrate-test.db")
	st, err := iamstore.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	n, err := RunSQLiteUp(st, migrationsDir)
	require.NoError(t, err)
	require.Positive(t, n)

	require.NoError(t, AssertRequired(context.Background(), st, runtimeconfig.RequiredMigration, dbPath))

	n2, err := RunSQLiteUp(st, migrationsDir)
	require.NoError(t, err)
	require.Equal(t, 0, n2)
}
