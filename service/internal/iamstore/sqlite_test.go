package iamstore

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenCreatesDatabaseWithoutCGO(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "iam.sqlite")

	st, err := Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	require.NoError(t, st.Ping(t.Context()))
}

func TestOpenMemoryWithoutCGO(t *testing.T) {
	st, err := OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	require.NoError(t, st.Ping(t.Context()))
}
