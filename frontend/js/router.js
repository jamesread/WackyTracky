import { createRouter, createWebHistory } from 'vue-router';
import { HomeIcon } from '@hugeicons/core-free-icons';
import { PinIcon } from '@hugeicons/core-free-icons';
import { Settings01Icon } from '@hugeicons/core-free-icons';
import { DashboardSquareSettingIcon } from '@hugeicons/core-free-icons';
import { UserShield01Icon } from '@hugeicons/core-free-icons';

import Welcome from '../resources/vue/views/Welcome.vue';
import ListView from '../resources/vue/views/ListView.vue';
import NavOptions from '../resources/vue/components/NavOptions.vue';
import Options from '../resources/vue/views/Options.vue';
import Diagnostics from '../resources/vue/views/Diagnostics.vue';
import TaskPropertyProperties from '../resources/vue/views/TaskPropertyProperties.vue';
import ControlPanel from '../resources/vue/views/ControlPanel.vue';
import IamHub from '../resources/vue/views/IamHub.vue';
import UsersAdmin from '../resources/vue/views/UsersAdmin.vue';
import UserInfoAdmin from '../resources/vue/views/UserInfoAdmin.vue';
import UserGroupsAdmin from '../resources/vue/views/UserGroupsAdmin.vue';
import UserGroupEdit from '../resources/vue/views/UserGroupEdit.vue';
import RbacRolesAdmin from '../resources/vue/views/RbacRolesAdmin.vue';
import RbacRoleEdit from '../resources/vue/views/RbacRoleEdit.vue';
import RbacPermissionsAdmin from '../resources/vue/views/RbacPermissionsAdmin.vue';
import MyPermissions from '../resources/vue/views/MyPermissions.vue';
import ApiKeysAdmin from '../resources/vue/views/ApiKeysAdmin.vue';
import UserControlPanel from '../resources/vue/views/UserControlPanel.vue';
import ChangePassword from '../resources/vue/views/ChangePassword.vue';
import { fetchAppStatus } from './modules/appStatus.js';
import {
	canAccessControlPanelFromStatus,
	canAccessIamFromStatus,
} from './modules/rbacAccess.js';

const routes = [
	{
		name: 'Welcome',
		path: '/',
		component: Welcome,
		meta: {
			title: 'Welcome',
			icon: HomeIcon,
			description: 'Choose a list to view or add tasks.',
		},
	},
	{
		name: 'ListView',
		path: '/lists/:listId',
		component: ListView,
		props: true,
		meta: {
			title: 'List',
			icon: PinIcon,
			breadcrumbs: (route) => {
				return [
					{ name: 'Home', href: '/' },
					{ name: 'List: ' + String(route.params.listId), href: '/lists/' + route.params.listId }
				];
			}
		}
	},
	{
		name: 'Search',
		path: '/search',
		component: ListView,
		props: (route) => ({ listId: null, searchQuery: route.query.q || '' }),
		meta: {
			title: 'Search',
			icon: PinIcon,
			description: 'Search tasks across all lists by text, tags, or contexts.',
			breadcrumbs: (route) => {
				return [
					{ name: 'Home', to: '/' },
					{ name: 'Search: ' + (route.query.q || ''), to: '/search' }
				];
			}
		}
	},
	{
		name: 'NavOptions',
		path: '/options',
		component: NavOptions,
		meta: {
			title: 'Options',
			icon: Settings01Icon,
			description: 'Shortcuts to settings, task properties, and diagnostics.',
		}
	},
	{
		name: 'TaskPropertyProperties',
		path: '/options/task-property-properties',
		component: TaskPropertyProperties,
		meta: {
			title: 'TPPs',
			icon: Settings01Icon,
			description: 'Configure tags and contexts (colors, order).',
		}
	},
	{
		name: 'Settings',
		path: '/settings',
		component: Options,
		meta: {
			title: 'Settings',
			icon: Settings01Icon,
			description: 'Server and client settings.',
		}
	},
	{
		name: 'Diagnostics',
		path: '/diagnostics',
		component: Diagnostics,
		meta: {
			title: 'Diagnostics',
			icon: Settings01Icon,
			description: 'Debug and health information.',
		}
	},
	{
		name: 'userControlPanel',
		path: '/user-control-panel',
		component: UserControlPanel,
		meta: {
			title: 'User Control Panel',
			requiresAuth: true,
		},
	},
	{
		name: 'changePassword',
		path: '/change-password',
		component: ChangePassword,
		meta: {
			title: 'Change password',
			requiresAuth: true,
		},
	},
	{
		name: 'myPermissions',
		path: '/my-permissions',
		component: MyPermissions,
		meta: {
			title: 'My Permissions',
			requiresAuth: true,
		},
	},
	{
		name: 'apiKeys',
		path: '/api-keys',
		component: ApiKeysAdmin,
		meta: {
			title: 'API keys',
			requiresAuth: true,
		},
	},
	{
		name: 'controlPanel',
		path: '/control-panel',
		component: ControlPanel,
		meta: {
			title: 'System Control Panel',
			icon: DashboardSquareSettingIcon,
			description: 'System and app administration.',
			requiresAuth: true,
			requiresControlPanel: true,
		},
	},
	{
		name: 'iam',
		path: '/control-panel/iam',
		component: IamHub,
		meta: {
			title: 'IAM',
			icon: UserShield01Icon,
			description: 'Users, groups, and roles.',
			requiresAuth: true,
			requiresIam: true,
		},
	},
	{
		name: 'users',
		path: '/users',
		component: UsersAdmin,
		meta: { title: 'Users', requiresAuth: true, requiresIam: true },
	},
	{
		name: 'userInfo',
		path: '/users/:id',
		component: UserInfoAdmin,
		meta: { title: 'User', requiresAuth: true, requiresIam: true },
	},
	{
		name: 'user-groups',
		path: '/user-groups',
		component: UserGroupsAdmin,
		meta: { title: 'User groups', requiresAuth: true, requiresIam: true },
	},
	{
		name: 'userGroupEdit',
		path: '/user-groups/:id',
		component: UserGroupEdit,
		meta: { title: 'User group', requiresAuth: true, requiresIam: true },
	},
	{
		name: 'rbac-roles',
		path: '/settings/rbac',
		component: RbacRolesAdmin,
		meta: { title: 'Roles', requiresAuth: true, requiresIam: true },
	},
	{
		name: 'rbac-permissions',
		path: '/settings/rbac/permissions',
		component: RbacPermissionsAdmin,
		meta: { title: 'Permissions', requiresAuth: true, requiresIam: true },
	},
	{
		name: 'rbac-role',
		path: '/settings/rbac/:id',
		component: RbacRoleEdit,
		meta: { title: 'Role', requiresAuth: true, requiresIam: true },
	},
	{ path: '/iam', redirect: { name: 'iam' } },
	{ path: '/user-control-panel/permissions', redirect: { name: 'myPermissions' } },
]

const router = createRouter({
	history: createWebHistory(),
	routes,
});

router.beforeEach(async (to) => {
	let st
	try {
		st = await fetchAppStatus()
	} catch {
		if (to.meta.requiresAuth) {
			return { name: 'Welcome' }
		}
		return
	}
	if (to.meta.requiresAuth && !st?.isLoggedIn) {
		return { name: 'Welcome' }
	}
	if (to.meta.requiresControlPanel && !canAccessControlPanelFromStatus(st)) {
		return '/'
	}
	if (to.meta.requiresIam && !canAccessIamFromStatus(st)) {
		return '/'
	}
})

export default router;
