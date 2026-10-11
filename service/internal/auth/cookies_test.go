package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionCookieName(t *testing.T) {
	require.Equal(t, "wacky-tracky-sid", SessionCookieName())
}

func TestSecureCookiesEnabled(t *testing.T) {
	t.Setenv("WACKYTRACKY_SECURE_COOKIES", "")
	require.True(t, SecureCookiesEnabled())
	t.Setenv("WACKYTRACKY_SECURE_COOKIES", "false")
	require.False(t, SecureCookiesEnabled())
}

func TestNewSessionCookie(t *testing.T) {
	t.Setenv("WACKYTRACKY_SECURE_COOKIES", "false")
	c := NewSessionCookie("abc")
	require.Equal(t, SessionCookieName(), c.Name)
	require.Equal(t, "abc", c.Value)
	require.True(t, c.HttpOnly)
	require.Equal(t, "/", c.Path)
	require.Equal(t, http.SameSiteStrictMode, c.SameSite)
	require.False(t, c.Secure)
}

func TestClearSessionCookie(t *testing.T) {
	c := ClearSessionCookie()
	require.Equal(t, -1, c.MaxAge)
	require.Empty(t, c.Value)
}
