package auth

import (
	"net/http"
	"strings"
)

// SSOOptOutCookieName marks a browser that logged out of a JWT (eg Cloudflare Access)
// session. While it is set, JWTs are ignored so the login form can be used with a
// local account. It is readable by the SPA (not HttpOnly), which clears it to sign
// in with SSO again; it carries no credentials.
const SSOOptOutCookieName = "wacky-tracky-sso-off"

const ssoOptOutMaxAge = 365 * 24 * 60 * 60

// SSOOptOut lets a user log out of a proxy JWT session. armature-iam checks JWTs
// before the session cookie, so without it the next request signs them straight back in.
type SSOOptOut struct {
	header string
	cookie string
}

// NewSSOOptOut reads where the JWT comes from (auth.jwt.header / cookieName).
func NewSSOOptOut(authYAML map[string]any) *SSOOptOut {
	jwt, _ := authYAML["jwt"].(map[string]any)
	header, _ := jwt["header"].(string)
	cookie, _ := jwt["cookieName"].(string)
	return &SSOOptOut{header: strings.TrimSpace(header), cookie: strings.TrimSpace(cookie)}
}

func (s *SSOOptOut) configured() bool {
	return s != nil && (s.header != "" || s.cookie != "")
}

// present reports whether the request carries a JWT from a configured source.
func (s *SSOOptOut) present(r *http.Request) bool {
	if !s.configured() {
		return false
	}
	return s.inHeader(r) || s.inCookie(r)
}

func (s *SSOOptOut) inHeader(r *http.Request) bool {
	return s.header != "" && r.Header.Get(s.header) != ""
}

func (s *SSOOptOut) inCookie(r *http.Request) bool {
	if s.cookie == "" {
		return false
	}
	c, err := r.Cookie(s.cookie)
	return err == nil && c.Value != ""
}

func hasSSOOptOut(r *http.Request) bool {
	c, err := r.Cookie(SSOOptOutCookieName)
	return err == nil && c.Value != ""
}

// StripJWTIfOptedOut removes the JWT header and cookie when the browser has opted out
// of SSO, so the auth layer falls through to the local session (or no user).
func (s *SSOOptOut) StripJWTIfOptedOut(r *http.Request) *http.Request {
	if !s.present(r) || !hasSSOOptOut(r) {
		return r
	}
	clone := r.Clone(r.Context())
	if s.header != "" {
		clone.Header.Del(s.header)
	}
	if s.cookie != "" {
		s.dropJWTCookie(clone, r.Cookies())
	}
	return clone
}

// dropJWTCookie rewrites the Cookie header of r without the JWT cookie.
func (s *SSOOptOut) dropJWTCookie(r *http.Request, cookies []*http.Cookie) {
	kept := make([]string, 0, len(cookies))
	for _, c := range cookies {
		if c.Name != s.cookie {
			kept = append(kept, c.Name+"="+c.Value)
		}
	}
	r.Header.Del("Cookie")
	if len(kept) > 0 {
		r.Header.Set("Cookie", strings.Join(kept, "; "))
	}
}

// Wrap must sit outside the auth layer's WrapHandler, so the JWT is gone before it is checked.
func (s *SSOOptOut) Wrap(in http.Handler) http.Handler {
	if !s.configured() {
		return in
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		in.ServeHTTP(w, s.StripJWTIfOptedOut(r))
	})
}

// Cookie returns the cookie to set on logout when the request was made with a
// configured JWT, so the user is not signed straight back in by the proxy.
func (s *SSOOptOut) Cookie(header http.Header) (http.Cookie, bool) {
	if !s.present(&http.Request{Header: header}) {
		return http.Cookie{}, false
	}
	return http.Cookie{
		Name:     SSOOptOutCookieName,
		Value:    "1",
		Path:     "/",
		MaxAge:   ssoOptOutMaxAge,
		SameSite: http.SameSiteLaxMode,
		Secure:   SecureCookiesEnabled(),
	}, true
}
