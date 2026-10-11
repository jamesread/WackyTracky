package migrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindSQLiteMigrationsDir(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	repo := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	t.Chdir(repo)

	dir, err := FindSQLiteMigrationsDir()
	require.NoError(t, err)
	require.DirExists(t, dir)
	require.FileExists(t, filepath.Join(dir, "1.iam.sql"))
}
