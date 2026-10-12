package clientapi

import (
	"context"
	"testing"

	"connectrpc.com/authn"
	"connectrpc.com/connect"
	"github.com/jamesread/armature-iam/password"
	"github.com/jamesread/armature-iam/rbac"
	iamstoretypes "github.com/jamesread/armature-iam/store"
	"github.com/stretchr/testify/require"

	pb "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1"
	"github.com/wacky-tracky/wacky-tracky-server/internal/auth"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/dummy"
	"github.com/wacky-tracky/wacky-tracky-server/internal/iamstore"
	"github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

func TestBootstrapIAM_CreatesAdmin(t *testing.T) {
	st, err := iamstore.OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	api := GetNewClientAPI(&dummy.Dummy{}, st)
	require.NoError(t, api.BootstrapIAM(context.Background()))

	count, err := st.CountUserAccounts(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, count)

	user, err := st.GetUserByUsername(context.Background(), "admin")
	require.NoError(t, err)
	require.NotNil(t, user)
	ok, err := password.Verify(user.PasswordHash, "admin")
	require.NoError(t, err)
	require.True(t, ok)

	rbac, err := st.LoadEffectiveRBAC(context.Background(), user.ID)
	require.NoError(t, err)
	require.True(t, rbac.IsSuperuser)
}

func TestBootstrapIAM_Idempotent(t *testing.T) {
	st, err := iamstore.OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	api := GetNewClientAPI(&dummy.Dummy{}, st)
	require.NoError(t, api.BootstrapIAM(context.Background()))
	require.NoError(t, api.BootstrapIAM(context.Background()))

	count, err := st.CountUserAccounts(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestGetStatus_DevDisableAuthReportsLoggedIn(t *testing.T) {
	t.Setenv("WACKYTRACKY_DEV_DISABLE_AUTH", "true")
	require.True(t, auth.DevDisableAuth())

	st, err := iamstore.OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	api := GetNewClientAPI(&dummy.Dummy{}, st)
	au := &auth.AuthenticatedUser{
		User: &iamstoretypes.UserAccountRow{ID: 0, Username: "anonymous"},
		RBAC: &rbac.EffectiveRBAC{IsSuperuser: true, Permissions: map[string]bool{}},
	}
	ctx := authn.SetInfo(context.Background(), au)
	res, err := api.GetStatus(ctx, connect.NewRequest(&pb.GetStatusRequest{}))
	require.NoError(t, err)
	require.True(t, res.Msg.IsLoggedIn)
	require.True(t, res.Msg.RbacIsSuperuser)
	require.Equal(t, "anonymous", res.Msg.Username)
}

func TestGetStatus_DatabaseDriverForSystemStatus(t *testing.T) {
	orig := runtimeconfig.RuntimeConfig.Database.Driver
	runtimeconfig.RuntimeConfig.Database.Driver = "todotxt"
	t.Cleanup(func() { runtimeconfig.RuntimeConfig.Database.Driver = orig })

	st, err := iamstore.OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	api := GetNewClientAPI(&dummy.Dummy{}, st)

	t.Run("superuser", func(t *testing.T) {
		au := &auth.AuthenticatedUser{
			User: &iamstoretypes.UserAccountRow{ID: 1, Username: "admin"},
			RBAC: &rbac.EffectiveRBAC{IsSuperuser: true, Permissions: map[string]bool{}},
		}
		ctx := authn.SetInfo(context.Background(), au)
		res, err := api.GetStatus(ctx, connect.NewRequest(&pb.GetStatusRequest{}))
		require.NoError(t, err)
		require.Equal(t, "todotxt", res.Msg.DatabaseDriver)
	})

	t.Run("system.logs", func(t *testing.T) {
		au := &auth.AuthenticatedUser{
			User: &iamstoretypes.UserAccountRow{ID: 2, Username: "ops"},
			RBAC: &rbac.EffectiveRBAC{Permissions: map[string]bool{rbac.PermissionSystemLogs: true}},
		}
		ctx := authn.SetInfo(context.Background(), au)
		res, err := api.GetStatus(ctx, connect.NewRequest(&pb.GetStatusRequest{}))
		require.NoError(t, err)
		require.Equal(t, "todotxt", res.Msg.DatabaseDriver)
	})

	t.Run("app.access only", func(t *testing.T) {
		au := &auth.AuthenticatedUser{
			User: &iamstoretypes.UserAccountRow{ID: 3, Username: "member"},
			RBAC: &rbac.EffectiveRBAC{Permissions: map[string]bool{rbac.PermissionAppAccess: true}},
		}
		ctx := authn.SetInfo(context.Background(), au)
		res, err := api.GetStatus(ctx, connect.NewRequest(&pb.GetStatusRequest{}))
		require.NoError(t, err)
		require.Empty(t, res.Msg.DatabaseDriver)
	})

	t.Run("anonymous", func(t *testing.T) {
		res, err := api.GetStatus(context.Background(), connect.NewRequest(&pb.GetStatusRequest{}))
		require.NoError(t, err)
		require.Empty(t, res.Msg.DatabaseDriver)
	})
}
