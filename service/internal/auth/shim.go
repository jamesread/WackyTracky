package auth

import (
	"context"
	"net/http"

	authlib "github.com/jamesread/httpauthshim"
	"github.com/jamesread/httpauthshim/authpublic"
	"github.com/jamesread/httpauthshim/sessions"
	log "github.com/sirupsen/logrus"
)

type contextKey struct{}

// UserFromRequest returns the authenticated user from the request context, or nil if not set.
func UserFromRequest(r *http.Request) *authpublic.AuthenticatedUser {
	u := r.Context().Value(contextKey{})
	if u == nil {
		return nil
	}
	return u.(*authpublic.AuthenticatedUser)
}

// NewAuthShim creates an AuthShimContext with the given config and YAML session persistence.
// Caller must call Shutdown() when done.
func NewAuthShim(cfg *authpublic.Config) (*authlib.AuthShimContext, error) {
	if cfg == nil {
		cfg = &authpublic.Config{}
	}
	sessionStorage := sessions.NewSessionStorage(sessions.NewYAMLPersistence())
	ctx, err := authlib.NewAuthShimContext(cfg, sessionStorage)
	if err != nil {
		return nil, err
	}
	log.Info("HTTP auth shim initialized")
	return ctx, nil
}

// Middleware wraps next with authentication. It runs the auth chain on each request,
// stores the user in the request context, and if requireAuth is true returns 401 for guests.
func Middleware(authCtx *authlib.AuthShimContext, requireAuth bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := authCtx.AuthFromHttpReq(r)
		ctx := context.WithValue(r.Context(), contextKey{}, user)
		if requireAuth && user.IsGuest() {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
