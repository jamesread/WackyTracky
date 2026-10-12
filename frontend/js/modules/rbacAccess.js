/** Permissions that unlock the System Control Panel when IAM (or settings) is implemented. */
export const CONTROL_PANEL_PERMISSIONS = [
	'users.view',
	'usergroups.view',
	'rbac.view',
	'system.settings',
	'system.logs',
]

export const IAM_PERMISSIONS = ['users.view', 'usergroups.view', 'rbac.view']

/**
 * @param {{ isLoggedIn?: boolean, rbacIsSuperuser?: boolean, rbacPermissions?: readonly string[] } | null | undefined} st
 * @param {string} name
 */
export function hasPermission(st, name) {
	return Boolean(st?.rbacIsSuperuser) || (st?.rbacPermissions || []).includes(name)
}

/**
 * @param {{ isLoggedIn?: boolean, rbacIsSuperuser?: boolean, rbacPermissions?: readonly string[] } | null | undefined} st
 */
export function canAccessControlPanelFromStatus(st) {
	if (!st?.isLoggedIn) return false
	if (st.rbacIsSuperuser) return true
	return CONTROL_PANEL_PERMISSIONS.some((p) => (st.rbacPermissions || []).includes(p))
}

/**
 * @param {{ isLoggedIn?: boolean, rbacIsSuperuser?: boolean, rbacPermissions?: readonly string[] } | null | undefined} st
 */
export function canAccessIamFromStatus(st) {
	if (!st?.isLoggedIn) return false
	if (st.rbacIsSuperuser) return true
	return IAM_PERMISSIONS.some((p) => (st.rbacPermissions || []).includes(p))
}

/**
 * @param {{ isLoggedIn?: boolean, rbacIsSuperuser?: boolean, rbacPermissions?: readonly string[] } | null | undefined} st
 */
export function canAccessSettingsFromStatus(st) {
	if (!st?.isLoggedIn) return false
	if (st.rbacIsSuperuser) return true
	return (st.rbacPermissions || []).includes('system.settings')
}

/** Matches GetStatus fields reserved for ops/admin (e.g. database_driver). */
export const SYSTEM_STATUS_PERMISSIONS = [
	'system.logs',
	...CONTROL_PANEL_PERMISSIONS,
]

/**
 * @param {{ isLoggedIn?: boolean, rbacIsSuperuser?: boolean, rbacPermissions?: readonly string[] } | null | undefined} st
 */
export function canViewSystemStatusFromStatus(st) {
	if (!st?.isLoggedIn) return false
	if (st.rbacIsSuperuser) return true
	const perms = st.rbacPermissions || []
	return SYSTEM_STATUS_PERMISSIONS.some((p) => perms.includes(p))
}
