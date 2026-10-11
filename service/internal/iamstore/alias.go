package iamstore

import iamstorepkg "github.com/jamesread/armature-iam/store"

const (
	UserCreatedByAdmin = iamstorepkg.UserCreatedByAdmin
	UserCreatedBySSO   = iamstorepkg.UserCreatedBySSO
)

type (
	Store                = iamstorepkg.Store
	UserAccountRow       = iamstorepkg.UserAccountRow
	SessionRow           = iamstorepkg.SessionRow
	APIKeyRow            = iamstorepkg.APIKeyRow
	UserGroupRow         = iamstorepkg.UserGroupRow
	RBACPermissionRow    = iamstorepkg.RBACPermissionRow
	RBACRoleRow          = iamstorepkg.RBACRoleRow
	MyPermissionAuditRow = iamstorepkg.MyPermissionAuditRow
)
