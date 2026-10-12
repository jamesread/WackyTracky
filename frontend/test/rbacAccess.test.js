import assert from 'node:assert/strict'
import { test } from 'node:test'
import {
	canAccessControlPanelFromStatus,
	canAccessIamFromStatus,
	canAccessSettingsFromStatus,
	canViewSystemStatusFromStatus,
	hasPermission,
} from '../js/modules/rbacAccess.js'

test('control panel is hidden when logged out', () => {
	assert.equal(canAccessControlPanelFromStatus({ isLoggedIn: false, rbacIsSuperuser: true }), false)
})

test('control panel is hidden without privilege', () => {
	assert.equal(
		canAccessControlPanelFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: false,
			rbacPermissions: [],
		}),
		false,
	)
})

test('superuser can open the control panel', () => {
	assert.equal(
		canAccessControlPanelFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: true,
			rbacPermissions: [],
		}),
		true,
	)
})

test('IAM view permission unlocks the control panel', () => {
	assert.equal(
		canAccessControlPanelFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: false,
			rbacPermissions: ['users.view'],
		}),
		true,
	)
})

test('settings permission unlocks the control panel', () => {
	assert.equal(
		canAccessControlPanelFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: false,
			rbacPermissions: ['system.settings'],
		}),
		true,
	)
})

test('system.impersonate alone does not unlock the control panel', () => {
	assert.equal(
		canAccessControlPanelFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: false,
			rbacPermissions: ['system.impersonate'],
		}),
		false,
	)
})

test('system.logs unlocks the control panel for quick status', () => {
	assert.equal(
		canAccessControlPanelFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: false,
			rbacPermissions: ['system.logs'],
		}),
		true,
	)
})

test('IAM tile follows IAM privileges', () => {
	assert.equal(
		canAccessIamFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: false,
			rbacPermissions: ['usergroups.view'],
		}),
		true,
	)
	assert.equal(
		canAccessIamFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: false,
			rbacPermissions: ['system.settings'],
		}),
		false,
	)
})

test('settings tile follows system.settings', () => {
	assert.equal(
		canAccessSettingsFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: false,
			rbacPermissions: ['system.settings'],
		}),
		true,
	)
	assert.equal(
		canAccessSettingsFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: false,
			rbacPermissions: ['users.view'],
		}),
		false,
	)
})

test('hasPermission grants superuser everything', () => {
	assert.equal(
		hasPermission({ isLoggedIn: true, rbacIsSuperuser: true, rbacPermissions: [] }, 'users.delete'),
		true,
	)
	assert.equal(
		hasPermission(
			{ isLoggedIn: true, rbacIsSuperuser: false, rbacPermissions: ['users.view'] },
			'users.delete',
		),
		false,
	)
})

test('system status is hidden when logged out', () => {
	assert.equal(canViewSystemStatusFromStatus({ isLoggedIn: false, rbacIsSuperuser: true }), false)
})

test('system.logs unlocks system status', () => {
	assert.equal(
		canViewSystemStatusFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: false,
			rbacPermissions: ['system.logs'],
		}),
		true,
	)
})

test('control panel privileges unlock system status', () => {
	assert.equal(
		canViewSystemStatusFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: false,
			rbacPermissions: ['users.view'],
		}),
		true,
	)
})

test('ordinary members cannot view system status', () => {
	assert.equal(
		canViewSystemStatusFromStatus({
			isLoggedIn: true,
			rbacIsSuperuser: false,
			rbacPermissions: ['app.access'],
		}),
		false,
	)
})
