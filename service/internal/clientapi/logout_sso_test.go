package clientapi

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1"
	"github.com/wacky-tracky/wacky-tracky-server/internal/auth"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/dummy"
	"github.com/wacky-tracky/wacky-tracky-server/internal/iamstore"
)

func logoutSetCookies(t *testing.T, api *wackyTrackyClientService, header http.Header) []string {
	t.Helper()
	req := connect.NewRequest(&pb.LogoutRequest{})
	for k, v := range header {
		req.Header()[k] = v
	}
	res, err := api.Logout(context.Background(), req)
	require.NoError(t, err)
	return res.Header().Values("Set-Cookie")
}

func hasCookie(cookies []string, name string) bool {
	for _, c := range cookies {
		if strings.HasPrefix(c, name+"=") {
			return true
		}
	}
	return false
}

func TestLogout_SetsSSOOptOutForJWTSessions(t *testing.T) {
	st, err := iamstore.OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	api := GetNewClientAPI(&dummy.Dummy{}, st)
	api.ssoOptOut = auth.NewSSOOptOut(map[string]any{
		"jwt": map[string]any{"header": "Cf-Access-Jwt-Assertion"},
	})

	withJWT := http.Header{}
	withJWT.Set("Cf-Access-Jwt-Assertion", "token")
	cookies := logoutSetCookies(t, api, withJWT)
	assert.True(t, hasCookie(cookies, auth.SessionCookieName()), "session cookie is cleared")
	assert.True(t, hasCookie(cookies, auth.SSOOptOutCookieName), "JWT session opts out of SSO")

	cookies = logoutSetCookies(t, api, http.Header{})
	assert.False(t, hasCookie(cookies, auth.SSOOptOutCookieName), "no JWT, no opt-out")
}

func TestLogout_NoSSOConfigured(t *testing.T) {
	st, err := iamstore.OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	api := GetNewClientAPI(&dummy.Dummy{}, st)
	withJWT := http.Header{}
	withJWT.Set("Cf-Access-Jwt-Assertion", "token")
	cookies := logoutSetCookies(t, api, withJWT)
	assert.False(t, hasCookie(cookies, auth.SSOOptOutCookieName))
}
