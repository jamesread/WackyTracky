package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	iamstoretypes "github.com/jamesread/armature-iam/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1/clientv1connect"
	"github.com/wacky-tracky/wacky-tracky-server/internal/iamstore"
)

const (
	testJWTSecret = "test-hmac-secret"
	testJWTHeader = "Cf-Access-Jwt-Assertion"
	testJWTCookie = "CF_Authorization"
	testJWTIssuer = "https://team.cloudflareaccess.com"
	testJWTAud    = "test-aud"
)

func testJWTAuthYAML() map[string]any {
	return map[string]any{
		"jwt": map[string]any{
			"hmacSecret":    testJWTSecret,
			"issuer":        testJWTIssuer,
			"aud":           testJWTAud,
			"claimUsername": "email",
			"header":        testJWTHeader,
			"cookieName":    testJWTCookie,
		},
	}
}

func signTestJWT(t *testing.T, email string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"iss":   testJWTIssuer,
		"aud":   testJWTAud,
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	s, err := tok.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)
	return s
}

func newJWTRequest(t *testing.T, email string) *http.Request {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), "POST", clientv1connect.WackyTrackyClientServiceGetStatusProcedure, nil)
	tok := signTestJWT(t, email)
	req.Header.Set(testJWTHeader, tok)
	req.AddCookie(&http.Cookie{Name: testJWTCookie, Value: tok})
	return req
}

func newJWTLayer(t *testing.T) (*Layer, *iamstore.SQLite) {
	t.Helper()
	st, err := iamstore.OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	require.NoError(t, st.EnsureRBACBootstrap(context.Background()))
	return NewLayer(st, testJWTAuthYAML(), nil), st
}

func TestJWTAutoCreatesUser(t *testing.T) {
	layer, st := newJWTLayer(t)
	ctx := context.Background()

	info, err := layer.Handle(ctx, newJWTRequest(t, "alice@example.com"))
	require.NoError(t, err)
	au, ok := info.(*AuthenticatedUser)
	require.True(t, ok, "JWT request should be authenticated")
	assert.Equal(t, "alice@example.com", au.User.Username)

	user, err := st.GetUserByUsername(ctx, "alice@example.com")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, iamstoretypes.UserCreatedBySSO, user.CreatedBy)
}

func TestStripJWTIfOptedOut(t *testing.T) {
	s := NewSSOOptOut(testJWTAuthYAML())

	req := newJWTRequest(t, "alice@example.com")
	assert.Same(t, req, s.StripJWTIfOptedOut(req), "no opt-out cookie: request unchanged")

	req.AddCookie(&http.Cookie{Name: SSOOptOutCookieName, Value: "1"})
	req.AddCookie(&http.Cookie{Name: SessionCookieName(), Value: "local-sid"})
	stripped := s.StripJWTIfOptedOut(req)

	assert.Empty(t, stripped.Header.Get(testJWTHeader))
	_, err := stripped.Cookie(testJWTCookie)
	require.ErrorIs(t, err, http.ErrNoCookie)
	sid, err := stripped.Cookie(SessionCookieName())
	require.NoError(t, err)
	assert.Equal(t, "local-sid", sid.Value, "local session cookie is kept")
	assert.NotEmpty(t, req.Header.Get(testJWTHeader), "original request is not modified")
}

func TestSSOOptOut_NotConfigured(t *testing.T) {
	var nilOptOut *SSOOptOut
	for _, s := range []*SSOOptOut{nilOptOut, NewSSOOptOut(nil), NewSSOOptOut(map[string]any{})} {
		req := newJWTRequest(t, "alice@example.com")
		req.AddCookie(&http.Cookie{Name: SSOOptOutCookieName, Value: "1"})
		assert.Same(t, req, s.StripJWTIfOptedOut(req))
		_, ok := s.Cookie(req.Header)
		assert.False(t, ok)
	}
}

func TestSSOOptOut_IgnoresJWTEndToEnd(t *testing.T) {
	layer, _ := newJWTLayer(t)
	s := NewSSOOptOut(testJWTAuthYAML())

	var called bool
	var gotUser *AuthenticatedUser
	handler := s.Wrap(layer.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		gotUser = UserFromContext(r.Context())
	})))

	handler.ServeHTTP(httptest.NewRecorder(), newJWTRequest(t, "alice@example.com"))
	require.True(t, called)
	require.NotNil(t, gotUser, "without the opt-out cookie the JWT signs in")
	assert.Equal(t, "alice@example.com", gotUser.User.Username)

	called, gotUser = false, nil
	req := newJWTRequest(t, "alice@example.com")
	req.AddCookie(&http.Cookie{Name: SSOOptOutCookieName, Value: "1"})
	handler.ServeHTTP(httptest.NewRecorder(), req)
	require.True(t, called, "GetStatus is allowed unauthenticated")
	assert.Nil(t, gotUser, "opted-out browser is not signed in by the JWT")
}

func TestSSOOptOutCookie(t *testing.T) {
	s := NewSSOOptOut(testJWTAuthYAML())

	c, ok := s.Cookie(newJWTRequest(t, "alice@example.com").Header)
	require.True(t, ok)
	assert.Equal(t, SSOOptOutCookieName, c.Name)
	assert.Equal(t, "/", c.Path)
	assert.Positive(t, c.MaxAge)
	assert.False(t, c.HttpOnly, "the SPA clears it to sign in with SSO again")

	_, ok = s.Cookie(http.Header{})
	assert.False(t, ok, "no JWT on the request: nothing to opt out of")
}
