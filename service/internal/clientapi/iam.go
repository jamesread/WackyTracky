package clientapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/jamesread/armature-iam/password"
	iamstorepkg "github.com/jamesread/armature-iam/store"
	pb "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1"
	"github.com/wacky-tracky/wacky-tracky-server/internal/auth"
	"github.com/wacky-tracky/wacky-tracky-server/internal/iamstore"
)

func toProtoRbacPermission(p *iamstore.RBACPermissionRow) *pb.RbacPermission {
	if p == nil {
		return nil
	}
	return &pb.RbacPermission{
		Id:          int32(p.ID),
		Name:        p.Name,
		Description: p.Description,
	}
}

func toProtoRbacRole(r *iamstore.RBACRoleRow) *pb.RbacRole {
	if r == nil {
		return nil
	}
	return &pb.RbacRole{
		Id:            int32(r.ID),
		Name:          r.Name,
		Description:   r.Description,
		PermissionIds: int32Slice(r.PermissionIDs),
		GroupCount:    int32(r.GroupCount),
		UserCount:     int32(r.UserCount),
	}
}

func toProtoUserGroup(g *iamstore.UserGroupRow) *pb.UserGroup {
	if g == nil {
		return nil
	}
	return &pb.UserGroup{
		Id:          int32(g.ID),
		Name:        g.Name,
		MemberCount: int32(g.MemberCount),
	}
}

func toProtoApiKey(k *iamstore.APIKeyRow) *pb.ApiKey {
	if k == nil {
		return nil
	}
	return &pb.ApiKey{
		Id:         int32(k.ID),
		Name:       k.Name,
		CreatedAt:  k.CreatedAt,
		LastUsedAt: k.LastUsedAt,
		ReadOnly:   k.ReadOnly,
	}
}

func toProtoMyPermissionAuditRow(r *iamstore.MyPermissionAuditRow) *pb.MyPermissionAuditRow {
	if r == nil {
		return nil
	}
	return &pb.MyPermissionAuditRow{
		Permission:     r.Permission,
		Granted:        r.Granted,
		GrantingGroups: append([]string(nil), r.GrantingGroups...),
	}
}

func int32Slice(ids []int) []int32 {
	out := make([]int32, len(ids))
	for i, id := range ids {
		out[i] = int32(id)
	}
	return out
}

func intSlice(ids []int32) []int {
	out := make([]int, len(ids))
	for i, id := range ids {
		out[i] = int(id)
	}
	return out
}

var failedPreconditionStoreErrs = []error{
	iamstorepkg.ErrNoSuperuser,
	iamstorepkg.ErrSystemRole,
	iamstorepkg.ErrSystemGroup,
	iamstorepkg.ErrRenameSystemRole,
}

var failedPreconditionMsgFragments = []string{
	"cannot delete system",
	"cannot modify system",
	"cannot rename system",
	"cannot set permissions for system",
	"refusing to leave the system without a superuser",
}

func mapStoreError(err error) *connect.Error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return connect.NewError(connect.CodeNotFound, err)
	}
	if isFailedPreconditionStoreErr(err) {
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}
	if strings.Contains(err.Error(), "already exists") {
		return connect.NewError(connect.CodeAlreadyExists, err)
	}
	return connect.NewError(connect.CodeInternal, err)
}

func isFailedPreconditionStoreErr(err error) bool {
	for _, target := range failedPreconditionStoreErrs {
		if errors.Is(err, target) {
			return true
		}
	}
	msg := err.Error()
	for _, fragment := range failedPreconditionMsgFragments {
		if strings.Contains(msg, fragment) {
			return true
		}
	}
	return false
}

func (api *wackyTrackyClientService) impersonatorUserID(ctx context.Context, header http.Header, au *auth.AuthenticatedUser) (int, error) {
	if sid := sessionIDFromHeader(header); sid != "" {
		sess, err := api.iamStore.GetSessionBySID(ctx, sid)
		if err != nil {
			return 0, connect.NewError(connect.CodeInternal, err)
		}
		if sess != nil && sess.ImpersonatorUserID != nil {
			return *sess.ImpersonatorUserID, nil
		}
	}
	return au.User.ID, nil
}

func (api *wackyTrackyClientService) userGroupsForUser(ctx context.Context, userID int) ([]*pb.UserGroup, error) {
	groupIDs, err := api.iamStore.ListUserGroupIDsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	groups := make([]*pb.UserGroup, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		group, err := api.iamStore.GetUserGroupByID(ctx, groupID)
		if err != nil || group == nil {
			continue
		}
		groups = append(groups, toProtoUserGroup(group))
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})
	return groups, nil
}

func (api *wackyTrackyClientService) GetUsers(ctx context.Context, _ *connect.Request[pb.GetUsersRequest]) (*connect.Response[pb.GetUsersResponse], error) {
	if _, err := api.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := api.iamStore.ListUserAccounts(ctx)
	if err != nil {
		api.log.WithError(err).Error("list user accounts")
		return nil, mapStoreError(err)
	}
	out := &pb.GetUsersResponse{}
	for i := range rows {
		user := toProtoUser(&rows[i])
		groups, err := api.userGroupsForUser(ctx, rows[i].ID)
		if err != nil {
			api.log.WithError(err).Error("list user groups for user")
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		user.UserGroups = groups
		out.Users = append(out.Users, user)
	}
	return connect.NewResponse(out), nil
}

func (api *wackyTrackyClientService) GetUser(ctx context.Context, req *connect.Request[pb.GetUserRequest]) (*connect.Response[pb.GetUserResponse], error) {
	if _, err := api.requireAuth(ctx); err != nil {
		return nil, err
	}
	out, err := api.getUserResponse(ctx, int(req.Msg.UserId))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return nil, err
		}
		api.log.WithError(err).Error("get user")
		return nil, err
	}
	return connect.NewResponse(out), nil
}

func (api *wackyTrackyClientService) getUserResponse(ctx context.Context, userID int) (*pb.GetUserResponse, error) {
	user, err := api.iamStore.GetUserByID(ctx, userID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	if user == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
	}
	groups, err := api.userGroupsForUser(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return &pb.GetUserResponse{User: toProtoUser(user), UserGroups: groups}, nil
}

func (api *wackyTrackyClientService) CreateUser(ctx context.Context, req *connect.Request[pb.CreateUserRequest]) (*connect.Response[pb.CreateUserResponse], error) {
	if _, err := api.requireWrite(ctx); err != nil {
		return nil, err
	}
	username, hash, err := createUserCredentials(req.Msg)
	if err != nil {
		return nil, err
	}
	user, err := api.createUserAccount(ctx, username, hash)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&pb.CreateUserResponse{
		StandardResponse: &pb.StandardResponse{Success: true, Message: "User created"},
		User:             toProtoUser(user),
	}), nil
}

func createUserCredentials(msg *pb.CreateUserRequest) (string, string, error) {
	username := strings.TrimSpace(msg.Username)
	if username == "" {
		return "", "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("username required"))
	}
	hash, err := hashOptionalPassword(msg.Password)
	if err != nil {
		return "", "", err
	}
	return username, hash, nil
}

func hashOptionalPassword(pwd string) (string, error) {
	if pwd == "" {
		return "", nil
	}
	if len(pwd) < 8 {
		return "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("password must be at least 8 characters"))
	}
	hash, err := password.Hash(pwd)
	if err != nil {
		return "", connect.NewError(connect.CodeInternal, err)
	}
	return hash, nil
}

func (api *wackyTrackyClientService) createUserAccount(ctx context.Context, username, hash string) (*iamstore.UserAccountRow, error) {
	id, err := api.iamStore.CreateUserAccount(ctx, username, hash, iamstore.UserCreatedByAdmin)
	if err != nil {
		api.log.WithError(err).Error("create user account")
		return nil, mapStoreError(err)
	}
	if err := api.finalizeCreatedUser(ctx, id); err != nil {
		return nil, err
	}
	user, err := api.iamStore.GetUserByID(ctx, id)
	if err != nil {
		api.log.WithError(err).Error("get created user")
		return nil, mapStoreError(err)
	}
	return user, nil
}

func (api *wackyTrackyClientService) finalizeCreatedUser(ctx context.Context, id int) error {
	if err := api.iamStore.EnsureUserInEveryoneGroup(ctx, id); err != nil {
		api.log.WithError(err).Error("ensure user in everyone group")
		return mapStoreError(err)
	}
	if err := api.iamStore.EnsureRBACBootstrap(ctx); err != nil {
		api.log.WithError(err).Error("ensure rbac bootstrap")
		return mapStoreError(err)
	}
	return nil
}

func (api *wackyTrackyClientService) DeleteUser(ctx context.Context, req *connect.Request[pb.DeleteUserRequest]) (*connect.Response[pb.DeleteUserResponse], error) {
	au, err := api.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	userID := int(req.Msg.UserId)
	if userID == au.User.ID {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("cannot delete your own account"))
	}
	if err := api.iamStore.DeleteUserAccount(ctx, userID); err != nil {
		api.log.WithError(err).Error("delete user account")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&pb.DeleteUserResponse{
		StandardResponse: &pb.StandardResponse{Success: true, Message: "User deleted"},
	}), nil
}

func (api *wackyTrackyClientService) ResetUserPassword(ctx context.Context, req *connect.Request[pb.ResetUserPasswordRequest]) (*connect.Response[pb.ResetUserPasswordResponse], error) {
	if _, err := api.requireWrite(ctx); err != nil {
		return nil, err
	}
	if len(req.Msg.NewPassword) < 8 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("new password must be at least 8 characters"))
	}
	hash, err := password.Hash(req.Msg.NewPassword)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := api.iamStore.UpdateUserPassword(ctx, int(req.Msg.UserId), hash); err != nil {
		api.log.WithError(err).Error("reset user password")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&pb.ResetUserPasswordResponse{
		StandardResponse: &pb.StandardResponse{Success: true, Message: "Password reset"},
	}), nil
}

func (api *wackyTrackyClientService) ListRbacPermissions(ctx context.Context, _ *connect.Request[pb.ListRbacPermissionsRequest]) (*connect.Response[pb.ListRbacPermissionsResponse], error) {
	if _, err := api.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := api.iamStore.ListRBACPermissions(ctx)
	if err != nil {
		api.log.WithError(err).Error("list rbac permissions")
		return nil, mapStoreError(err)
	}
	out := &pb.ListRbacPermissionsResponse{}
	for i := range rows {
		out.Permissions = append(out.Permissions, toProtoRbacPermission(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (api *wackyTrackyClientService) ListRbacRoles(ctx context.Context, _ *connect.Request[pb.ListRbacRolesRequest]) (*connect.Response[pb.ListRbacRolesResponse], error) {
	if _, err := api.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := api.iamStore.ListRBACRoles(ctx)
	if err != nil {
		api.log.WithError(err).Error("list rbac roles")
		return nil, mapStoreError(err)
	}
	out := &pb.ListRbacRolesResponse{}
	for i := range rows {
		out.Roles = append(out.Roles, toProtoRbacRole(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (api *wackyTrackyClientService) CreateRbacRole(ctx context.Context, req *connect.Request[pb.CreateRbacRoleRequest]) (*connect.Response[pb.CreateRbacRoleResponse], error) {
	if _, err := api.requireWrite(ctx); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Msg.Name)
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name required"))
	}
	id, err := api.iamStore.CreateRBACRole(ctx, name, req.Msg.Description, intSlice(req.Msg.PermissionIds))
	if err != nil {
		api.log.WithError(err).Error("create rbac role")
		return nil, mapStoreError(err)
	}
	role, err := api.iamStore.GetRBACRole(ctx, id)
	if err != nil {
		api.log.WithError(err).Error("get created rbac role")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&pb.CreateRbacRoleResponse{Role: toProtoRbacRole(role)}), nil
}

func (api *wackyTrackyClientService) UpdateRbacRole(ctx context.Context, req *connect.Request[pb.UpdateRbacRoleRequest]) (*connect.Response[pb.RbacRole], error) {
	if _, err := api.requireWrite(ctx); err != nil {
		return nil, err
	}
	if err := api.iamStore.UpdateRBACRole(ctx, int(req.Msg.Id), req.Msg.Name, req.Msg.Description, intSlice(req.Msg.PermissionIds)); err != nil {
		api.log.WithError(err).Error("update rbac role")
		return nil, mapStoreError(err)
	}
	role, err := api.iamStore.GetRBACRole(ctx, int(req.Msg.Id))
	if err != nil {
		api.log.WithError(err).Error("get updated rbac role")
		return nil, mapStoreError(err)
	}
	if role == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("role not found"))
	}
	return connect.NewResponse(toProtoRbacRole(role)), nil
}

func (api *wackyTrackyClientService) DeleteRbacRole(ctx context.Context, req *connect.Request[pb.DeleteRbacRoleRequest]) (*connect.Response[pb.DeleteRbacRoleResponse], error) {
	if _, err := api.requireWrite(ctx); err != nil {
		return nil, err
	}
	if err := api.iamStore.DeleteRBACRole(ctx, int(req.Msg.Id)); err != nil {
		api.log.WithError(err).Error("delete rbac role")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&pb.DeleteRbacRoleResponse{}), nil
}

func (api *wackyTrackyClientService) GetUserRbacRoles(ctx context.Context, req *connect.Request[pb.GetUserRbacRolesRequest]) (*connect.Response[pb.GetUserRbacRolesResponse], error) {
	if _, err := api.requireAuth(ctx); err != nil {
		return nil, err
	}
	names, err := api.iamStore.GetUserRbacRoleNames(ctx, int(req.Msg.UserId))
	if err != nil {
		api.log.WithError(err).Error("get user rbac role names")
		return nil, mapStoreError(err)
	}
	roles, err := api.rbacRolesByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&pb.GetUserRbacRolesResponse{Roles: roles}), nil
}

func (api *wackyTrackyClientService) rbacRolesByNames(ctx context.Context, names []string) ([]*pb.RbacRole, error) {
	if len(names) == 0 {
		return nil, nil
	}
	all, err := api.iamStore.ListRBACRoles(ctx)
	if err != nil {
		api.log.WithError(err).Error("list rbac roles")
		return nil, mapStoreError(err)
	}
	return filterRolesByNames(all, names), nil
}

func filterRolesByNames(all []iamstore.RBACRoleRow, names []string) []*pb.RbacRole {
	want := make(map[string]struct{}, len(names))
	for _, name := range names {
		want[name] = struct{}{}
	}
	out := make([]*pb.RbacRole, 0, len(names))
	for i := range all {
		if _, ok := want[all[i].Name]; ok {
			out = append(out, toProtoRbacRole(&all[i]))
		}
	}
	return out
}

func (api *wackyTrackyClientService) GetUserGroupRbacRoles(ctx context.Context, req *connect.Request[pb.GetUserGroupRbacRolesRequest]) (*connect.Response[pb.GetUserGroupRbacRolesResponse], error) {
	if _, err := api.requireAuth(ctx); err != nil {
		return nil, err
	}
	roleIDs, err := api.iamStore.GetUserGroupRbacRoleIDs(ctx, int(req.Msg.GroupId))
	if err != nil {
		api.log.WithError(err).Error("get user group rbac role ids")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&pb.GetUserGroupRbacRolesResponse{
		RoleIds: int32Slice(roleIDs),
	}), nil
}

func (api *wackyTrackyClientService) SetUserGroupRbacRoles(ctx context.Context, req *connect.Request[pb.SetUserGroupRbacRolesRequest]) (*connect.Response[pb.SetUserGroupRbacRolesResponse], error) {
	if _, err := api.requireWrite(ctx); err != nil {
		return nil, err
	}
	if err := api.iamStore.SetUserGroupRbacRoles(ctx, int(req.Msg.GroupId), intSlice(req.Msg.RoleIds)); err != nil {
		api.log.WithError(err).Error("set user group rbac roles")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&pb.SetUserGroupRbacRolesResponse{
		StandardResponse: &pb.StandardResponse{Success: true, Message: "Group roles updated"},
	}), nil
}

func (api *wackyTrackyClientService) GetRbacRoleUsers(ctx context.Context, req *connect.Request[pb.GetRbacRoleUsersRequest]) (*connect.Response[pb.GetRbacRoleUsersResponse], error) {
	if _, err := api.requireAuth(ctx); err != nil {
		return nil, err
	}
	usernames, err := api.iamStore.ListRbacRoleUsernames(ctx, int(req.Msg.RoleId))
	if err != nil {
		api.log.WithError(err).Error("list rbac role usernames")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&pb.GetRbacRoleUsersResponse{
		Usernames: append([]string(nil), usernames...),
	}), nil
}

func (api *wackyTrackyClientService) GetRbacRoleGroups(ctx context.Context, req *connect.Request[pb.GetRbacRoleGroupsRequest]) (*connect.Response[pb.GetRbacRoleGroupsResponse], error) {
	if _, err := api.requireAuth(ctx); err != nil {
		return nil, err
	}
	groupNames, err := api.iamStore.ListRbacRoleGroupNames(ctx, int(req.Msg.RoleId))
	if err != nil {
		api.log.WithError(err).Error("list rbac role group names")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&pb.GetRbacRoleGroupsResponse{
		GroupNames: append([]string(nil), groupNames...),
	}), nil
}

func (api *wackyTrackyClientService) GetMyPermissionsAudit(ctx context.Context, _ *connect.Request[pb.GetMyPermissionsAuditRequest]) (*connect.Response[pb.GetMyPermissionsAuditResponse], error) {
	au, err := api.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	groupNames, roleNames, isSuperuser, rows, err := api.iamStore.GetMyPermissionsAudit(ctx, au.User.ID)
	if err != nil {
		api.log.WithError(err).Error("get my permissions audit")
		return nil, mapStoreError(err)
	}
	out := &pb.GetMyPermissionsAuditResponse{
		GroupNames:  append([]string(nil), groupNames...),
		RoleNames:   append([]string(nil), roleNames...),
		IsSuperuser: isSuperuser,
		Permissions: make([]*pb.MyPermissionAuditRow, 0, len(rows)),
	}
	for i := range rows {
		out.Permissions = append(out.Permissions, toProtoMyPermissionAuditRow(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (api *wackyTrackyClientService) ListUserGroups(ctx context.Context, _ *connect.Request[pb.ListUserGroupsRequest]) (*connect.Response[pb.ListUserGroupsResponse], error) {
	if _, err := api.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := api.iamStore.ListUserGroups(ctx)
	if err != nil {
		api.log.WithError(err).Error("list user groups")
		return nil, mapStoreError(err)
	}
	out := &pb.ListUserGroupsResponse{}
	for i := range rows {
		out.Groups = append(out.Groups, toProtoUserGroup(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (api *wackyTrackyClientService) CreateUserGroup(ctx context.Context, req *connect.Request[pb.CreateUserGroupRequest]) (*connect.Response[pb.CreateUserGroupResponse], error) {
	if _, err := api.requireWrite(ctx); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Msg.Name)
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name required"))
	}
	id, err := api.iamStore.CreateUserGroup(ctx, name)
	if err != nil {
		api.log.WithError(err).Error("create user group")
		return nil, mapStoreError(err)
	}
	group, err := api.iamStore.GetUserGroupByID(ctx, id)
	if err != nil {
		api.log.WithError(err).Error("get created user group")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&pb.CreateUserGroupResponse{Group: toProtoUserGroup(group)}), nil
}

func (api *wackyTrackyClientService) DeleteUserGroup(ctx context.Context, req *connect.Request[pb.DeleteUserGroupRequest]) (*connect.Response[pb.DeleteUserGroupResponse], error) {
	if _, err := api.requireWrite(ctx); err != nil {
		return nil, err
	}
	if err := api.iamStore.DeleteUserGroup(ctx, int(req.Msg.GroupId)); err != nil {
		api.log.WithError(err).Error("delete user group")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&pb.DeleteUserGroupResponse{}), nil
}

func (api *wackyTrackyClientService) GetUserGroupMembers(ctx context.Context, req *connect.Request[pb.GetUserGroupMembersRequest]) (*connect.Response[pb.GetUserGroupMembersResponse], error) {
	if _, err := api.requireAuth(ctx); err != nil {
		return nil, err
	}
	memberIDs, err := api.iamStore.ListUserGroupMemberIDs(ctx, int(req.Msg.GroupId))
	if err != nil {
		api.log.WithError(err).Error("list user group member ids")
		return nil, mapStoreError(err)
	}
	members, err := api.loadUsersByIDs(ctx, memberIDs)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&pb.GetUserGroupMembersResponse{Members: members}), nil
}

func (api *wackyTrackyClientService) loadUsersByIDs(ctx context.Context, ids []int) ([]*pb.UserAccount, error) {
	out := make([]*pb.UserAccount, 0, len(ids))
	for _, userID := range ids {
		user, err := api.iamStore.GetUserByID(ctx, userID)
		if err != nil {
			api.log.WithError(err).Error("get user group member")
			return nil, mapStoreError(err)
		}
		if user != nil {
			out = append(out, toProtoUser(user))
		}
	}
	return out, nil
}

func (api *wackyTrackyClientService) SetUserGroupMembers(ctx context.Context, req *connect.Request[pb.SetUserGroupMembersRequest]) (*connect.Response[pb.SetUserGroupMembersResponse], error) {
	if _, err := api.requireWrite(ctx); err != nil {
		return nil, err
	}
	if err := api.iamStore.SetUserGroupMembers(ctx, int(req.Msg.GroupId), intSlice(req.Msg.UserIds)); err != nil {
		api.log.WithError(err).Error("set user group members")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&pb.SetUserGroupMembersResponse{
		StandardResponse: &pb.StandardResponse{Success: true, Message: "Group members updated"},
	}), nil
}

func (api *wackyTrackyClientService) ImpersonateUser(ctx context.Context, req *connect.Request[pb.ImpersonateUserRequest]) (*connect.Response[pb.ImpersonateUserResponse], error) {
	au, err := api.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	targetID := int(req.Msg.UserId)
	if err := api.validateImpersonationTarget(ctx, targetID); err != nil {
		return nil, err
	}
	impersonatorID, err := api.impersonatorUserID(ctx, req.Header(), au)
	if err != nil {
		return nil, err
	}
	if impersonatorID == targetID {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("cannot impersonate yourself"))
	}
	return api.startImpersonationSession(ctx, req.Header(), targetID, impersonatorID)
}

func (api *wackyTrackyClientService) validateImpersonationTarget(ctx context.Context, targetID int) error {
	target, err := api.iamStore.GetUserByID(ctx, targetID)
	if err != nil {
		api.log.WithError(err).Error("get impersonation target")
		return mapStoreError(err)
	}
	if target == nil {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
	}
	return nil
}

func (api *wackyTrackyClientService) startImpersonationSession(ctx context.Context, header http.Header, targetID, impersonatorID int) (*connect.Response[pb.ImpersonateUserResponse], error) {
	if oldSID := sessionIDFromHeader(header); oldSID != "" {
		_ = api.iamStore.DeleteSession(ctx, oldSID)
	}
	newSID := uuid.NewString()
	if err := api.iamStore.CreateSession(ctx, newSID, targetID, &impersonatorID); err != nil {
		api.log.WithError(err).Error("create impersonation session")
		return nil, mapStoreError(err)
	}
	res := connect.NewResponse(&pb.ImpersonateUserResponse{
		StandardResponse: &pb.StandardResponse{Success: true, Message: "Impersonation started"},
	})
	c := auth.NewSessionCookie(newSID)
	res.Header().Add("Set-Cookie", c.String())
	return res, nil
}

func (api *wackyTrackyClientService) StopImpersonation(ctx context.Context, req *connect.Request[pb.StopImpersonationRequest]) (*connect.Response[pb.StopImpersonationResponse], error) {
	if _, err := api.requireAuth(ctx); err != nil {
		return nil, err
	}
	impersonatorID, err := api.impersonatorFromSession(ctx, req.Header())
	if err != nil {
		return nil, err
	}
	return api.restoreImpersonatorSession(ctx, sessionIDFromHeader(req.Header()), impersonatorID)
}

func (api *wackyTrackyClientService) impersonatorFromSession(ctx context.Context, header http.Header) (int, error) {
	sid := sessionIDFromHeader(header)
	if sid == "" {
		return 0, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("not impersonating"))
	}
	sess, err := api.iamStore.GetSessionBySID(ctx, sid)
	if err != nil {
		api.log.WithError(err).Error("get session for stop impersonation")
		return 0, mapStoreError(err)
	}
	if sess == nil || sess.ImpersonatorUserID == nil {
		return 0, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("not impersonating"))
	}
	return *sess.ImpersonatorUserID, nil
}

func (api *wackyTrackyClientService) restoreImpersonatorSession(ctx context.Context, oldSID string, impersonatorID int) (*connect.Response[pb.StopImpersonationResponse], error) {
	if err := api.iamStore.DeleteSession(ctx, oldSID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		api.log.WithError(err).Error("delete impersonation session")
		return nil, mapStoreError(err)
	}
	newSID := uuid.NewString()
	if err := api.iamStore.CreateSession(ctx, newSID, impersonatorID, nil); err != nil {
		api.log.WithError(err).Error("restore impersonator session")
		return nil, mapStoreError(err)
	}
	res := connect.NewResponse(&pb.StopImpersonationResponse{
		StandardResponse: &pb.StandardResponse{Success: true, Message: "Impersonation stopped"},
	})
	c := auth.NewSessionCookie(newSID)
	res.Header().Add("Set-Cookie", c.String())
	return res, nil
}

func (api *wackyTrackyClientService) ListApiKeys(ctx context.Context, _ *connect.Request[pb.ListApiKeysRequest]) (*connect.Response[pb.ListApiKeysResponse], error) {
	au, err := api.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := api.iamStore.ListAPIKeysForUser(ctx, au.User.ID)
	if err != nil {
		api.log.WithError(err).Error("list api keys")
		return nil, mapStoreError(err)
	}
	out := &pb.ListApiKeysResponse{}
	for i := range rows {
		out.Keys = append(out.Keys, toProtoApiKey(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (api *wackyTrackyClientService) CreateApiKey(ctx context.Context, req *connect.Request[pb.CreateApiKeyRequest]) (*connect.Response[pb.CreateApiKeyResponse], error) {
	au, err := api.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Msg.Name)
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name required"))
	}
	secret, created, err := api.createAPIKeyForUser(ctx, au.User.ID, name, req.Msg.ReadOnly)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&pb.CreateApiKeyResponse{
		Key:    toProtoApiKey(created),
		Secret: secret,
	}), nil
}

func (api *wackyTrackyClientService) createAPIKeyForUser(ctx context.Context, userID int, name string, readOnly bool) (string, *iamstore.APIKeyRow, error) {
	secret, err := password.GenerateAPIKey("wt_")
	if err != nil {
		return "", nil, connect.NewError(connect.CodeInternal, err)
	}
	id, err := api.iamStore.CreateAPIKey(ctx, userID, name, secret, readOnly)
	if err != nil {
		api.log.WithError(err).Error("create api key")
		return "", nil, mapStoreError(err)
	}
	created, err := api.findAPIKeyByID(ctx, userID, id)
	if err != nil {
		return "", nil, err
	}
	return secret, created, nil
}

func (api *wackyTrackyClientService) findAPIKeyByID(ctx context.Context, userID, id int) (*iamstore.APIKeyRow, error) {
	rows, err := api.iamStore.ListAPIKeysForUser(ctx, userID)
	if err != nil {
		api.log.WithError(err).Error("list api keys after create")
		return nil, mapStoreError(err)
	}
	for i := range rows {
		if rows[i].ID == id {
			return &rows[i], nil
		}
	}
	return nil, nil
}

func (api *wackyTrackyClientService) DeleteApiKey(ctx context.Context, req *connect.Request[pb.DeleteApiKeyRequest]) (*connect.Response[pb.DeleteApiKeyResponse], error) {
	au, err := api.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	if err := api.iamStore.DeleteAPIKey(ctx, int(req.Msg.Id), au.User.ID); err != nil {
		api.log.WithError(err).Error("delete api key")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&pb.DeleteApiKeyResponse{}), nil
}
