package auth

import (
	"github.com/jamesread/armature-iam/rbac"
	"github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1/clientv1connect"
)

var procedurePermissions = map[string]string{
	clientv1connect.WackyTrackyClientServiceGetUserProcedure:               rbac.PermissionUsersView,
	clientv1connect.WackyTrackyClientServiceCreateUserProcedure:            rbac.PermissionUsersCreate,
	clientv1connect.WackyTrackyClientServiceDeleteUserProcedure:            rbac.PermissionUsersDelete,
	clientv1connect.WackyTrackyClientServiceResetUserPasswordProcedure:     rbac.PermissionUsersResetPassword,
	clientv1connect.WackyTrackyClientServiceListRbacPermissionsProcedure:   rbac.PermissionRbacView,
	clientv1connect.WackyTrackyClientServiceListRbacRolesProcedure:         rbac.PermissionRbacView,
	clientv1connect.WackyTrackyClientServiceGetUserRbacRolesProcedure:      rbac.PermissionRbacView,
	clientv1connect.WackyTrackyClientServiceGetUserGroupRbacRolesProcedure: rbac.PermissionRbacView,
	clientv1connect.WackyTrackyClientServiceGetRbacRoleUsersProcedure:      rbac.PermissionRbacView,
	clientv1connect.WackyTrackyClientServiceGetRbacRoleGroupsProcedure:     rbac.PermissionRbacView,
	clientv1connect.WackyTrackyClientServiceCreateRbacRoleProcedure:        rbac.PermissionRbacManage,
	clientv1connect.WackyTrackyClientServiceUpdateRbacRoleProcedure:        rbac.PermissionRbacManage,
	clientv1connect.WackyTrackyClientServiceDeleteRbacRoleProcedure:        rbac.PermissionRbacManage,
	clientv1connect.WackyTrackyClientServiceSetUserGroupRbacRolesProcedure: rbac.PermissionRbacManage,
	clientv1connect.WackyTrackyClientServiceListUserGroupsProcedure:        rbac.PermissionUserGroupsView,
	clientv1connect.WackyTrackyClientServiceGetUserGroupMembersProcedure:   rbac.PermissionUserGroupsView,
	clientv1connect.WackyTrackyClientServiceCreateUserGroupProcedure:       rbac.PermissionUserGroupsManage,
	clientv1connect.WackyTrackyClientServiceDeleteUserGroupProcedure:       rbac.PermissionUserGroupsManage,
	clientv1connect.WackyTrackyClientServiceSetUserGroupMembersProcedure:   rbac.PermissionUserGroupsManage,
	clientv1connect.WackyTrackyClientServiceImpersonateUserProcedure:       rbac.PermissionSystemImpersonate,
	clientv1connect.WackyTrackyClientServiceStopImpersonationProcedure:     "",
}

func RequiredPermission(procedureName string) string {
	if perm, ok := procedurePermissions[procedureName]; ok {
		return perm
	}
	return rbac.PermissionAppAccess
}
