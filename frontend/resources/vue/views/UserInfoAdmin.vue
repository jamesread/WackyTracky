<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import CheckGroup from 'picocrank/vue/components/CheckGroup.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, UserIcon } from '@hugeicons/core-free-icons'
import { fetchAppStatus, useAppStatus } from '../../../js/modules/appStatus.js'
import { hasPermission } from '../../../js/modules/rbacAccess.js'

const route = useRoute()
const router = useRouter()
const status = useAppStatus()
const userId = computed(() => Number(route.params.id))

const user = ref(null)
const allGroups = ref([])
const memberGroupIds = ref([])
const effectiveRoles = ref([])
const loading = ref(true)
const error = ref('')
const groupsError = ref('')
const savingGroups = ref(false)
const deleteError = ref('')
const deleting = ref(false)
const confirmDeleteOpen = ref(false)
const resetDialog = ref(null)
const resetPassword = ref('')
const resetError = ref('')
const resetting = ref(false)
const impersonating = ref(false)
const impersonateError = ref('')

const canManageGroups = computed(() => hasPermission(status.status, 'usergroups.manage'))
const canDelete = computed(
	() =>
		hasPermission(status.status, 'users.delete') &&
		Boolean(user.value?.username) &&
		user.value?.username !== status.status?.username,
)
const canResetPassword = computed(() => hasPermission(status.status, 'users.reset-password'))
const canImpersonate = computed(
	() =>
		hasPermission(status.status, 'system.impersonate') &&
		Boolean(user.value?.username) &&
		user.value?.username !== status.status?.username,
)

function formatDate(value) {
	if (!value) return '—'
	const d = new Date(value)
	return Number.isNaN(d.getTime()) ? value : d.toLocaleString()
}

async function load() {
	loading.value = true
	error.value = ''
	try {
		const [userRes, groupsRes, rolesRes] = await Promise.all([
			window.client.getUser({ userId: userId.value }),
			window.client.listUserGroups({}),
			window.client.getUserRbacRoles({ userId: userId.value }).catch(() => ({ roles: [] })),
		])
		user.value = userRes.user || null
		allGroups.value = groupsRes.groups || []
		effectiveRoles.value = rolesRes.roles || rolesRes.roleNames || []
		const fromUser = (userRes.userGroups || user.value?.userGroups || []).map((g) => g.id)
		if (fromUser.length) {
			memberGroupIds.value = fromUser
		} else {
			const memberships = await Promise.all(
				allGroups.value.map(async (g) => {
					const m = await window.client.getUserGroupMembers({ groupId: g.id })
					const ids = (m.members || []).map((u) => u.id)
					return ids.includes(userId.value) ? g.id : null
				}),
			)
			memberGroupIds.value = memberships.filter((id) => id != null)
		}
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
		user.value = null
	} finally {
		loading.value = false
	}
}

async function syncMembership(groupId, userAccountId, shouldBeMember) {
	const m = await window.client.getUserGroupMembers({ groupId })
	const current = new Set((m.members || []).map((u) => u.id))
	const isMember = current.has(userAccountId)
	if (shouldBeMember === isMember) return
	if (shouldBeMember) current.add(userAccountId)
	else current.delete(userAccountId)
	await window.client.setUserGroupMembers({ groupId, userIds: [...current] })
}

async function saveGroups() {
	if (!user.value || !canManageGroups.value) return
	savingGroups.value = true
	groupsError.value = ''
	try {
		const desired = new Set(memberGroupIds.value)
		const uid = user.value.id
		await Promise.all(
			allGroups.value.map((g) => syncMembership(g.id, uid, desired.has(g.id))),
		)
		await load()
	} catch (e) {
		groupsError.value = e instanceof Error ? e.message : String(e)
	} finally {
		savingGroups.value = false
	}
}

function openResetDialog() {
	resetPassword.value = ''
	resetError.value = ''
	resetDialog.value?.showModal()
	nextTick(() => document.getElementById('reset-password-input')?.focus())
}

function closeResetDialog() {
	resetDialog.value?.close()
}

async function submitResetPassword() {
	if (!user.value) return
	if (resetPassword.value.length < 8) {
		resetError.value = 'Password must be at least 8 characters'
		return
	}
	resetting.value = true
	resetError.value = ''
	try {
		await window.client.resetUserPassword({
			userId: user.value.id,
			password: resetPassword.value,
		})
		closeResetDialog()
	} catch (e) {
		resetError.value = e instanceof Error ? e.message : String(e)
	} finally {
		resetting.value = false
	}
}

async function removeUser() {
	if (!user.value?.id) return
	deleting.value = true
	deleteError.value = ''
	try {
		await window.client.deleteUser({ userId: user.value.id })
		confirmDeleteOpen.value = false
		router.push({ name: 'users' })
	} catch (e) {
		deleteError.value = e instanceof Error ? e.message : String(e)
	} finally {
		deleting.value = false
	}
}

async function impersonate() {
	if (!user.value) return
	impersonating.value = true
	impersonateError.value = ''
	try {
		await window.client.impersonateUser({ userId: user.value.id })
		await fetchAppStatus({ force: true })
		router.push('/')
		window.location.reload()
	} catch (e) {
		impersonateError.value = e instanceof Error ? e.message : String(e)
	} finally {
		impersonating.value = false
	}
}

const roleLabels = computed(() => {
	const roles = effectiveRoles.value
	if (!roles.length) return []
	if (typeof roles[0] === 'string') return roles
	return roles.map((r) => r.name || String(r.id))
})

onMounted(load)
</script>

<template>
	<dialog ref="resetDialog" class="dialog" @close="resetPassword = ''; resetError = ''">
		<h2>Reset password</h2>
		<p>Set a new password for {{ user?.username }}.</p>
		<FormLayout @submit.prevent="submitResetPassword">
			<FormField label="New password" for="reset-password-input">
				<input
					id="reset-password-input"
					v-model="resetPassword"
					type="password"
					autocomplete="new-password"
					minlength="8"
					required
					:disabled="resetting"
				/>
			</FormField>
			<p v-if="resetError" class="inline-notification error">{{ resetError }}</p>
			<template #actions>
				<button type="button" class="neutral" :disabled="resetting" @click="closeResetDialog">Cancel</button>
				<button type="submit" class="good" :disabled="resetting">
					{{ resetting ? 'Saving…' : 'Reset password' }}
				</button>
			</template>
		</FormLayout>
	</dialog>

	<div
		v-if="confirmDeleteOpen"
		class="confirm-overlay"
		role="dialog"
		aria-modal="true"
		aria-labelledby="confirm-delete-title"
		@click.self="confirmDeleteOpen = false"
	>
		<div class="confirm-dialog">
			<h2 id="confirm-delete-title">Delete user</h2>
			<p>Delete user {{ user?.username }}? This cannot be undone.</p>
			<p v-if="deleteError" class="inline-notification error">{{ deleteError }}</p>
			<div class="confirm-actions">
				<button type="button" class="neutral" @click="confirmDeleteOpen = false">Cancel</button>
				<button type="button" class="bad" :disabled="deleting" @click="removeUser">
					{{ deleting ? 'Deleting…' : 'Delete' }}
				</button>
			</div>
		</div>
	</div>

	<Section :title="user?.username || 'User'" :icon="UserIcon" :padding="true">
		<template #toolbar>
			<RouterLink :to="{ name: 'users' }" class="button inline-icon neutral">
				<HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
				<span>Users</span>
			</RouterLink>
			<button type="button" class="inline-icon neutral" :disabled="loading" @click="load">
				Refresh
			</button>
			<button
				v-if="canResetPassword && user"
				type="button"
				class="neutral"
				:disabled="loading"
				@click="openResetDialog"
			>
				Reset password
			</button>
			<button
				v-if="canImpersonate && user"
				type="button"
				class="neutral"
				:disabled="impersonating"
				@click="impersonate"
			>
				{{ impersonating ? 'Impersonating…' : 'Impersonate' }}
			</button>
		</template>

		<p v-if="loading" class="subtle">Loading…</p>
		<p v-if="error" class="inline-notification error">{{ error }}</p>
		<p v-if="impersonateError" class="inline-notification error">{{ impersonateError }}</p>

		<template v-else-if="user">
			<dl class="account-info">
				<dt>Username</dt>
				<dd>{{ user.username }}</dd>
				<dt>Created by</dt>
				<dd>{{ user.createdBy || '—' }}</dd>
				<dt>Account created</dt>
				<dd>{{ formatDate(user.createdAt) }}</dd>
				<dt>Effective roles</dt>
				<dd>
					<span v-if="roleLabels.length">{{ roleLabels.join(', ') }}</span>
					<span v-else class="subtle">None (via groups)</span>
				</dd>
			</dl>
		</template>
	</Section>

	<Section
		v-if="user && !loading"
		title="User groups"
		subtitle="Membership grants roles assigned to those groups."
		:padding="true"
	>
		<FormLayout @submit.prevent="saveGroups">
			<FormField label="Groups" component-has-label>
				<CheckGroup
					v-model="memberGroupIds"
					name="user-groups"
					:options="allGroups.map((g) => ({ value: g.id, label: g.name }))"
					:disabled="!canManageGroups || savingGroups"
				/>
			</FormField>
			<p v-if="groupsError" class="inline-notification error">{{ groupsError }}</p>
			<p v-if="!canManageGroups" class="subtle">You can view groups but not change membership.</p>
			<template v-if="canManageGroups" #actions>
				<button type="submit" class="good" :disabled="savingGroups">
					{{ savingGroups ? 'Saving…' : 'Save groups' }}
				</button>
			</template>
		</FormLayout>
	</Section>

	<Section
		v-if="user && canDelete"
		title="Danger zone"
		subtitle="Delete this sign-in account"
		:padding="true"
	>
		<p class="inline-notification error">
			Deleting {{ user.username }} removes their account and sign-in access. This cannot be undone.
		</p>
		<div role="toolbar" class="danger-zone-actions">
			<button type="button" class="bad" @click="confirmDeleteOpen = true">Delete user</button>
		</div>
	</Section>
</template>

<style scoped>
.account-info {
	display: grid;
	grid-template-columns: 200px 1fr;
	column-gap: 1em;
	margin: 0;
}
.account-info dt {
	font-weight: bold;
}
.account-info dd {
	margin: 0;
}
.confirm-overlay {
	position: fixed;
	inset: 0;
	z-index: 1000;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(0, 0, 0, 0.35);
}
.confirm-dialog {
	background: var(--femtocrank-bg, #fff);
	padding: 1.25rem;
	border-radius: 0.35rem;
	max-width: 24rem;
	width: 90%;
}
.confirm-actions {
	display: flex;
	justify-content: flex-end;
	gap: 0.5rem;
	margin-top: 1rem;
}
</style>
