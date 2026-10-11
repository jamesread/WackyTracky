package auth

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/jamesread/armature-iam/layer"
	"github.com/jamesread/armature-iam/store"
	"github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1/clientv1connect"
)

type Layer = layer.Layer
type AuthenticatedUser = layer.AuthenticatedUser

// DevDisableAuth reports whether WACKYTRACKY_DEV_DISABLE_AUTH=true (anonymous superuser).
func DevDisableAuth() bool {
	return os.Getenv("WACKYTRACKY_DEV_DISABLE_AUTH") == "true"
}

func NewLayer(st store.Store, authYAML map[string]any, log *logrus.Logger) *layer.Layer {
	if log == nil {
		log = logrus.New()
	}
	l, err := layer.New(st, layer.Config{
		CookieName: SessionCookieName(),
		AuthYAML:   authYAML,
		Logger:     log,
		AllowUnauthenticated: []string{
			clientv1connect.WackyTrackyClientServiceLoginWithUsernameAndPasswordProcedure,
			clientv1connect.WackyTrackyClientServiceGetStatusProcedure,
		},
		RequiredPermission: RequiredPermission,
		DevDisableAuth:     DevDisableAuth(),
		APIKeyPrefix:       "wt_",
		SecureCookies:      SecureCookiesEnabled(),
	})
	if err != nil {
		log.Fatalf("armature-iam init: %v", err)
	}
	return l
}
