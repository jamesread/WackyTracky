<script setup>
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
	DashboardSquareSettingIcon,
	Key01Icon,
	UserShield01Icon,
} from '@hugeicons/core-free-icons'
import {
	fetchAppStatus,
	invalidateAppStatus,
	useAppStatus,
} from '../../../js/modules/appStatus.js'
import { canAccessControlPanelFromStatus } from '../../../js/modules/rbacAccess.js'

const router = useRouter()
const status = useAppStatus()
const navRef = ref(null)
const loggingOut = ref(false)
const error = ref('')
const stoppingImpersonation = ref(false)

const userData = computed(() => status.status)
const isImpersonating = computed(() => Boolean(status.status?.isImpersonating || status.status?.impersonatorUsername))

async function refreshUserData() {
	error.value = ''
	try {
		await fetchAppStatus({ force: true })
		await nextTick()
		setupQuickActions()
	} catch (e) {
		error.value = e instanceof Error ? e.message : 'Refresh failed'
	}
}

function setupQuickActions() {
	const nav = navRef.value
	if (!nav) return
	nav.clearNavigationLinks()
	nav.addCallback('Change Password', () => router.push({ name: 'changePassword' }), {
		icon: Key01Icon,
		name: 'change-password',
		description: 'Update your account password',
	})
	nav.addCallback('API Keys', () => router.push({ name: 'apiKeys' }), {
		icon: Key01Icon,
		name: 'api-keys',
		description: 'Manage Bearer tokens for automation',
	})
	nav.addCallback('My Permissions', () => router.push({ name: 'myPermissions' }), {
		icon: UserShield01Icon,
		name: 'my-permissions',
		description: 'Review groups, roles, and effective access',
	})
	if (canAccessControlPanelFromStatus(status.status)) {
		nav.addCallback('System Control Panel', () => router.push({ name: 'controlPanel' }), {
			name: 'control-panel',
			icon: DashboardSquareSettingIcon,
			description: 'System administration',
		})
	}
}

async function logout() {
	loggingOut.value = true
	try {
		await window.client.logout({})
		invalidateAppStatus()
		await router.push('/')
		window.location.reload()
	} finally {
		loggingOut.value = false
	}
}

async function stopImpersonation() {
	stoppingImpersonation.value = true
	try {
		await window.client.stopImpersonation({})
		await fetchAppStatus({ force: true })
		window.location.reload()
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
	} finally {
		stoppingImpersonation.value = false
	}
}

function formatDate(value) {
	if (!value) return '—'
	const d = new Date(value)
	return Number.isNaN(d.getTime()) ? value : d.toLocaleString()
}

onMounted(() => {
	void refreshUserData()
})
</script>

<template>
	<p v-if="error" class="inline-notification error">{{ error }}</p>

	<Section v-if="userData?.isLoggedIn" title="Identity" :padding="true">
		<template #toolbar>
			<button type="button" class="inline-icon neutral" aria-label="Refresh" @click="refreshUserData">
				Refresh
			</button>
		</template>

		<h3>{{ userData.username }}</h3>
		<dl class="account-info">
			<dt>Username</dt>
			<dd>{{ userData.username }}</dd>
			<dt v-if="userData.accountCreatedAt || userData.createdAt">Account created</dt>
			<dd v-if="userData.accountCreatedAt || userData.createdAt">
				{{ formatDate(userData.accountCreatedAt || userData.createdAt) }}
			</dd>
			<template v-if="isImpersonating">
				<dt>Impersonating</dt>
				<dd>
					Yes
					<button
						type="button"
						class="neutral small"
						:disabled="stoppingImpersonation"
						@click="stopImpersonation"
					>
						{{ stoppingImpersonation ? 'Stopping…' : 'Stop impersonation' }}
					</button>
				</dd>
			</template>
		</dl>
	</Section>

	<Section v-if="userData?.isLoggedIn" title="Quick actions" :padding="true">
		<Navigation ref="navRef">
			<NavigationGrid />
		</Navigation>
	</Section>

	<Section v-if="userData?.isLoggedIn" title="Session" :padding="true">
		<p>End your current session and return to the login page.</p>
		<button type="button" class="bad" :disabled="loggingOut" @click="logout">
			{{ loggingOut ? 'Signing out…' : 'Sign Out' }}
		</button>
	</Section>
</template>

<style scoped>
.account-info {
	display: grid;
	grid-template-columns: 200px 1fr;
	column-gap: 1em;
	margin: 1em 0 0;
}
.account-info dt {
	font-weight: bold;
}
.account-info dd {
	margin: 0;
}
</style>
