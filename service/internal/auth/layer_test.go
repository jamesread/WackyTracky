package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wacky-tracky/wacky-tracky-server/internal/iamstore"
)

func TestNewLayer_EmptyAuthYAML(t *testing.T) {
	st, err := iamstore.OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	t.Setenv("WACKYTRACKY_DEV_DISABLE_AUTH", "true")
	layer := NewLayer(st, nil, nil)
	require.NotNil(t, layer)
}
