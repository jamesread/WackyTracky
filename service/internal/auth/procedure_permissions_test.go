package auth

import (
	"testing"

	"github.com/jamesread/armature-iam/rbac"
	"github.com/stretchr/testify/require"
	"github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1/clientv1connect"
)

func TestRequiredPermission(t *testing.T) {
	require.Equal(t, rbac.PermissionUsersView,
		RequiredPermission(clientv1connect.WackyTrackyClientServiceGetUserProcedure))
	require.Equal(t, rbac.PermissionUsersCreate,
		RequiredPermission(clientv1connect.WackyTrackyClientServiceCreateUserProcedure))
	require.Equal(t, rbac.PermissionUsersDelete,
		RequiredPermission(clientv1connect.WackyTrackyClientServiceDeleteUserProcedure))
	require.Equal(t, rbac.PermissionUsersResetPassword,
		RequiredPermission(clientv1connect.WackyTrackyClientServiceResetUserPasswordProcedure))
	require.Equal(t, rbac.PermissionRbacView,
		RequiredPermission(clientv1connect.WackyTrackyClientServiceListRbacRolesProcedure))
	require.Equal(t, rbac.PermissionRbacManage,
		RequiredPermission(clientv1connect.WackyTrackyClientServiceCreateRbacRoleProcedure))
	require.Equal(t, rbac.PermissionUserGroupsView,
		RequiredPermission(clientv1connect.WackyTrackyClientServiceListUserGroupsProcedure))
	require.Equal(t, rbac.PermissionUserGroupsManage,
		RequiredPermission(clientv1connect.WackyTrackyClientServiceSetUserGroupMembersProcedure))
	require.Equal(t, rbac.PermissionSystemImpersonate,
		RequiredPermission(clientv1connect.WackyTrackyClientServiceImpersonateUserProcedure))
	require.Equal(t, "",
		RequiredPermission(clientv1connect.WackyTrackyClientServiceStopImpersonationProcedure))
	require.Equal(t, rbac.PermissionAppAccess,
		RequiredPermission(clientv1connect.WackyTrackyClientServiceListTasksProcedure))
	require.Equal(t, rbac.PermissionAppAccess,
		RequiredPermission(clientv1connect.WackyTrackyClientServiceGetUsersProcedure))
}
