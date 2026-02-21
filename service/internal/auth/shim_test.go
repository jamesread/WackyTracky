package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authpublic "github.com/jamesread/httpauthshim/authpublic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserFromRequest_NoUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	user := UserFromRequest(req)
	assert.Nil(t, user)
}

func TestUserFromRequest_WithUser(t *testing.T) {
	u := &authpublic.AuthenticatedUser{}
	ctx := context.WithValue(context.Background(), contextKey{}, u)
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	user := UserFromRequest(req)
	require.NotNil(t, user)
	assert.Same(t, u, user)
}
