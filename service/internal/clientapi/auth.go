package clientapi

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/jamesread/armature-iam/password"
	"github.com/jamesread/armature-iam/rbac"
	pb "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1"
	"github.com/wacky-tracky/wacky-tracky-server/internal/auth"
	"github.com/wacky-tracky/wacky-tracky-server/internal/iamstore"
	"github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

// Privileges that may see operational fields on GetStatus (e.g. task DB driver).
var systemStatusPermissions = []string{
	rbac.PermissionSystemLogs,
	rbac.PermissionSystemSettings,
	rbac.PermissionUsersView,
	rbac.PermissionUserGroupsView,
	rbac.PermissionRbacView,
}

func (api *wackyTrackyClientService) BootstrapIAM(ctx context.Context) error {
	if api.iamStore == nil {
		return fmt.Errorf("iam store not configured")
	}
	if err := api.iamStore.EnsureRBACBootstrap(ctx); err != nil {
		return err
	}
	return api.ensureDefaultAdmin(ctx)
}

func (api *wackyTrackyClientService) ensureDefaultAdmin(ctx context.Context) error {
	count, err := api.iamStore.CountUserAccounts(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return api.createDefaultAdmin(ctx)
}

func (api *wackyTrackyClientService) createDefaultAdmin(ctx context.Context) error {
	hash, err := password.Hash("admin")
	if err != nil {
		return err
	}
	id, err := api.iamStore.CreateUserAccount(ctx, "admin", hash, iamstore.UserCreatedByAdmin)
	if err != nil {
		return err
	}
	if err := api.iamStore.EnsureUserInEveryoneGroup(ctx, id); err != nil {
		return err
	}
	return api.iamStore.EnsureRBACBootstrap(ctx)
}

func (api *wackyTrackyClientService) GetStatus(ctx context.Context, req *connect.Request[pb.GetStatusRequest]) (*connect.Response[pb.GetStatusResponse], error) {
	au := auth.UserFromContext(ctx)
	res := &pb.GetStatusResponse{
		UsesSecureCookies: auth.SecureCookiesEnabled(),
		Username:          "<anonymous>",
	}
	if au == nil || au.User == nil {
		return connect.NewResponse(res), nil
	}
	api.populateLoggedInStatus(ctx, au, req, res)
	return connect.NewResponse(res), nil
}

func (api *wackyTrackyClientService) populateLoggedInStatus(ctx context.Context, au *auth.AuthenticatedUser, req *connect.Request[pb.GetStatusRequest], res *pb.GetStatusResponse) {
	// DevDisableAuth uses anonymous ID 0 with superuser RBAC; still report logged-in.
	res.IsLoggedIn = au.User.ID > 0 || auth.DevDisableAuth()
	res.Username = au.User.Username
	api.fillAccountCreatedAt(ctx, au.User.ID, res)
	fillRBACStatus(au, res)
	fillSystemStatus(au, res)
	if sid := sessionIDFromHeader(req.Header()); sid != "" {
		api.fillImpersonation(ctx, sid, res)
	}
}

func canViewSystemStatus(au *auth.AuthenticatedUser) bool {
	if au == nil || au.RBAC == nil {
		return false
	}
	for _, p := range systemStatusPermissions {
		if au.RBAC.Has(p) {
			return true
		}
	}
	return false
}

func fillSystemStatus(au *auth.AuthenticatedUser, res *pb.GetStatusResponse) {
	if !canViewSystemStatus(au) {
		return
	}
	if db := runtimeconfig.RuntimeConfig.Database; db != nil {
		res.DatabaseDriver = db.Driver
	}
}

func (api *wackyTrackyClientService) fillAccountCreatedAt(ctx context.Context, userID int, res *pb.GetStatusResponse) {
	if userID <= 0 {
		return
	}
	if user, uerr := api.iamStore.GetUserByID(ctx, userID); uerr == nil && user != nil {
		res.AccountCreatedAt = user.CreatedAt
	}
}

func fillRBACStatus(au *auth.AuthenticatedUser, res *pb.GetStatusResponse) {
	if au.RBAC == nil {
		return
	}
	res.RbacIsSuperuser = au.RBAC.IsSuperuser
	for p := range au.RBAC.Permissions {
		res.RbacPermissions = append(res.RbacPermissions, p)
	}
	sort.Strings(res.RbacPermissions)
}

func (api *wackyTrackyClientService) fillImpersonation(ctx context.Context, sid string, res *pb.GetStatusResponse) {
	sess, err := api.iamStore.GetSessionBySID(ctx, sid)
	if err != nil || sess == nil || sess.ImpersonatorUserID == nil {
		return
	}
	res.IsImpersonating = true
	if imp, _ := api.iamStore.GetUserByID(ctx, *sess.ImpersonatorUserID); imp != nil {
		res.ImpersonatorUsername = imp.Username
	}
}

func sessionIDFromHeader(h http.Header) string {
	for _, part := range strings.Split(h.Get("Cookie"), ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, auth.SessionCookieName()+"=") {
			return strings.TrimPrefix(part, auth.SessionCookieName()+"=")
		}
	}
	return ""
}

func (api *wackyTrackyClientService) LoginWithUsernameAndPassword(ctx context.Context, req *connect.Request[pb.LoginWithUsernameAndPasswordRequest]) (*connect.Response[pb.LoginWithUsernameAndPasswordResponse], error) {
	username := strings.TrimSpace(req.Msg.Username)
	user, err := api.iamStore.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res := connect.NewResponse(&pb.LoginWithUsernameAndPasswordResponse{
		StandardResponse: &pb.StandardResponse{Success: false, Message: "Invalid username or password"},
	})
	if !passwordMatches(user, req.Msg.Password) {
		return res, nil
	}
	if err := api.createLoginSession(ctx, res, user); err != nil {
		return nil, err
	}
	return res, nil
}

func passwordMatches(user *iamstore.UserAccountRow, pwd string) bool {
	if user == nil || user.PasswordHash == "" {
		return false
	}
	match, err := password.Verify(user.PasswordHash, pwd)
	return err == nil && match
}

func (api *wackyTrackyClientService) createLoginSession(ctx context.Context, res *connect.Response[pb.LoginWithUsernameAndPasswordResponse], user *iamstore.UserAccountRow) error {
	sid := uuid.NewString()
	if err := api.iamStore.CreateSession(ctx, sid, user.ID, nil); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	res.Msg.StandardResponse = &pb.StandardResponse{Success: true, Message: "Login successful"}
	res.Msg.Username = user.Username
	c := auth.NewSessionCookie(sid)
	res.Header().Add("Set-Cookie", c.String())
	return nil
}

func (api *wackyTrackyClientService) Logout(ctx context.Context, req *connect.Request[pb.LogoutRequest]) (*connect.Response[pb.LogoutResponse], error) {
	if sid := sessionIDFromHeader(req.Header()); sid != "" {
		_ = api.iamStore.DeleteSession(ctx, sid)
	}
	res := connect.NewResponse(&pb.LogoutResponse{
		StandardResponse: &pb.StandardResponse{Success: true, Message: "Logged out"},
	})
	c := auth.ClearSessionCookie()
	res.Header().Add("Set-Cookie", c.String())
	return res, nil
}

func (api *wackyTrackyClientService) ChangePassword(ctx context.Context, req *connect.Request[pb.ChangePasswordRequest]) (*connect.Response[pb.ChangePasswordResponse], error) {
	au, err := api.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	if len(req.Msg.NewPassword) < 8 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("new password must be at least 8 characters"))
	}
	hash, err := api.verifyAndHashPasswordChange(ctx, au.User.ID, req.Msg)
	if err != nil {
		return nil, err
	}
	if err := api.iamStore.UpdateUserPassword(ctx, au.User.ID, hash); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&pb.ChangePasswordResponse{
		StandardResponse: &pb.StandardResponse{Success: true, Message: "Password changed"},
	}), nil
}

func (api *wackyTrackyClientService) verifyAndHashPasswordChange(ctx context.Context, userID int, msg *pb.ChangePasswordRequest) (string, error) {
	user, err := api.iamStore.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return "", connect.NewError(connect.CodeInternal, err)
	}
	if err := verifyCurrentPassword(user, msg.CurrentPassword); err != nil {
		return "", err
	}
	hash, err := password.Hash(msg.NewPassword)
	if err != nil {
		return "", connect.NewError(connect.CodeInternal, err)
	}
	return hash, nil
}

func verifyCurrentPassword(user *iamstore.UserAccountRow, current string) error {
	if user.PasswordHash == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user has no password set"))
	}
	match, err := password.Verify(user.PasswordHash, current)
	if err != nil || !match {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid current password"))
	}
	return nil
}

func toProtoUser(u *iamstore.UserAccountRow) *pb.UserAccount {
	if u == nil {
		return nil
	}
	return &pb.UserAccount{
		Id:        int32(u.ID),
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
		CreatedBy: u.CreatedBy,
	}
}

func (api *wackyTrackyClientService) requireAuth(ctx context.Context) (*auth.AuthenticatedUser, error) {
	au := auth.UserFromContext(ctx)
	if au == nil || au.User == nil || au.User.ID == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authentication required"))
	}
	return au, nil
}

func (api *wackyTrackyClientService) requireWrite(ctx context.Context) (*auth.AuthenticatedUser, error) {
	au, err := api.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if au.ReadOnly {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("read-only API key"))
	}
	return au, nil
}
