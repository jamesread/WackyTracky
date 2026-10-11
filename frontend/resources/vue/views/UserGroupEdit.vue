<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import CheckGroup from 'picocrank/vue/components/CheckGroup.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, Key01Icon, UserGroupIcon, UserIcon } from '@hugeicons/core-free-icons'
import { useAppStatus } from '../../../js/modules/appStatus.js'
import { hasPermission } from '../../../js/modules/rbacAccess.js'

const SYSTEM_GROUPS = new Set(['Everyone', 'Administrators'])

const route = useRoute()
const router = useRouter()
const status = useAppStatus()
const groupId = computed(() => Number(route.params.id))

const group = ref(null)
const users = ref([])
const memberIds = ref([])
const roleIds = ref([])
const allRoles = ref([])
const loading = ref(true)
const error = ref('')
const membersError = ref('')
const rolesError = ref('')
const savingMembers = ref(false)
const savingRoles = ref(false)
const deleting = ref(false)
const deleteError = ref('')
const confirmDeleteOpen = ref(false)

const canManage = computed(() => hasPermission(status.status, 'usergroups.manage'))
const canManageRoles = computed(() => hasPermission(status.status, 'rbac.manage'))
const isSystemGroup = computed(() => SYSTEM_GROUPS.has(group.value?.name || ''))
const sectionTitle = computed(() => group.value?.name || 'User group')

const memberUsernames = computed(() =>
	memberIds.value
		.map((id) => users.value.find((user) => user.id === id)?.username)
		.filter(Boolean),
)

const assignedRoleNames = computed(() =>
	roleIds.value
		.map((id) => allRoles.value.find((role) => role.id === id)?.name)
		.filter(Boolean),
)

async function load() {
	loading.value = true
	error.value = ''
	try {
		const [groupsRes, usersRes, rolesRes] = await Promise.all([
			window.client.listUserGroups({}),
			window.client.getUsers({}),
			window.client.listRbacRoles({}),
		])
		group.value = (groupsRes.groups || []).find((g) => g.id === groupId.value) || null
		users.value = usersRes.users || []
		allRoles.value = (rolesRes.roles || []).map((role) => ({ id: role.id, name: role.name }))
		if (!group.value) {
			error.value = 'User group not found'
			return
		}
		const [members, roles] = await Promise.all([
			window.client.getUserGroupMembers({ groupId: groupId.value }),
			window.client.getUserGroupRbacRoles({ groupId: groupId.value }),
		])
		memberIds.value = (members.members || []).map((m) => m.id)
		roleIds.value = roles.roleIds || []
	} catch (e) {
		error.value = e instanceof Error ? e.message : 'Load failed'
		group.value = null
	} finally {
		loading.value = false
	}
}

async function saveMembers() {
	if (!group.value || !canManage.value) return
	savingMembers.value = true
	membersError.value = ''
	try {
		await window.client.setUserGroupMembers({ groupId: group.value.id, userIds: memberIds.value })
		router.push({ name: 'user-groups' })
	} catch (e) {
		membersError.value = e instanceof Error ? e.message : String(e)
	} finally {
		savingMembers.value = false
	}
}

async function saveRoles() {
	if (!group.value || !canManageRoles.value) return
	savingRoles.value = true
	rolesError.value = ''
	try {
		await window.client.setUserGroupRbacRoles({ groupId: group.value.id, roleIds: roleIds.value })
		router.push({ name: 'user-groups' })
	} catch (e) {
		rolesError.value = e instanceof Error ? e.message : String(e)
	} finally {
		savingRoles.value = false
	}
}

async function removeGroup() {
	if (!group.value) return
	deleting.value = true
	deleteError.value = ''
	try {
		await window.client.deleteUserGroup({ groupId: group.value.id })
		confirmDeleteOpen.value = false
		router.push({ name: 'user-groups' })
	} catch (e) {
		deleteError.value = e instanceof Error ? e.message : String(e)
	} finally {
		deleting.value = false
	}
}

onMounted(load)
</script>

<template>
	<div
		v-if="confirmDeleteOpen"
		class="confirm-overlay"
		role="dialog"
		aria-modal="true"
		aria-labelledby="confirm-delete-group-title"
		@click.self="confirmDeleteOpen = false"
	>
		<div class="confirm-dialog">
			<h2 id="confirm-delete-group-title">Delete user group</h2>
			<p>Delete user group "{{ group?.name }}"?</p>
			<p v-if="deleteError" class="inline-notification error">{{ deleteError }}</p>
			<div class="confirm-actions">
				<button type="button" class="neutral" @click="confirmDeleteOpen = false">Cancel</button>
				<button type="button" class="bad" :disabled="deleting" @click="removeGroup">
					{{ deleting ? 'Deleting…' : 'Delete' }}
				</button>
			</div>
		</div>
	</div>

	<Section :title="sectionTitle" :icon="UserGroupIcon" :padding="true">
		<template #toolbar>
			<RouterLink :to="{ name: 'user-groups' }" class="button inline-icon neutral">
				<HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
				<span>User groups</span>
			</RouterLink>
			<button type="button" class="inline-icon neutral" :disabled="loading" @click="load">
				Refresh
			</button>
		</template>

		<p v-if="loading" class="subtle">Loading…</p>
		<p v-if="error" class="inline-notification error">{{ error }}</p>
		<p v-if="group && isSystemGroup" class="inline-notification note">
			This is a system group (Everyone / Administrators). It cannot be renamed or deleted.
		</p>

		<dl v-if="group" class="group-info">
			<dt>Name</dt>
			<dd>
				{{ group.name }}
				<span v-if="isSystemGroup" class="tag">system group</span>
			</dd>
			<dt>Members</dt>
			<dd>
				{{ memberIds.length }}
				<span v-if="memberUsernames.length" class="subtle">
					— {{ memberUsernames.join(', ') }}
				</span>
				<span v-else class="subtle"> — no members yet</span>
			</dd>
			<dt>Roles</dt>
			<dd>
				<span v-if="assignedRoleNames.length">{{ assignedRoleNames.join(', ') }}</span>
				<span v-else class="subtle">No roles assigned</span>
			</dd>
		</dl>
	</Section>

	<Section
		v-if="group && !loading"
		title="Members"
		subtitle="Users who belong to this group."
		:icon="UserIcon"
		:padding="true"
	>
		<FormLayout @submit.prevent="saveMembers">
			<FormField label="Members" component-has-label>
				<CheckGroup
					v-model="memberIds"
					name="group-members"
					:options="users.map((u) => ({ value: u.id, label: u.username }))"
					:disabled="!canManage || savingMembers"
				/>
			</FormField>
			<p v-if="membersError" class="inline-notification error">{{ membersError }}</p>
			<p v-if="!canManage" class="subtle">You can view members but not change them.</p>
			<template v-if="canManage" #actions>
				<RouterLink :to="{ name: 'user-groups' }" class="button neutral">Cancel</RouterLink>
				<button type="submit" class="good" :disabled="savingMembers">
					{{ savingMembers ? 'Saving…' : 'Save members' }}
				</button>
			</template>
		</FormLayout>
	</Section>

	<Section
		v-if="group && !loading"
		title="Roles"
		subtitle="Roles assigned to this group grant permissions to its members."
		:icon="Key01Icon"
		:padding="true"
	>
		<FormLayout @submit.prevent="saveRoles">
			<FormField label="Roles" component-has-label>
				<CheckGroup
					v-model="roleIds"
					name="group-roles"
					:options="allRoles.map((r) => ({ value: r.id, label: r.name }))"
					:disabled="!canManageRoles || savingRoles"
				/>
			</FormField>
			<p v-if="rolesError" class="inline-notification error">{{ rolesError }}</p>
			<p v-if="!canManageRoles" class="subtle">You can view roles but not change them.</p>
			<template v-if="canManageRoles" #actions>
				<RouterLink :to="{ name: 'user-groups' }" class="button neutral">Cancel</RouterLink>
				<button type="submit" class="good" :disabled="savingRoles">
					{{ savingRoles ? 'Saving…' : 'Save roles' }}
				</button>
			</template>
		</FormLayout>
	</Section>

	<Section
		v-if="group && canManage && !isSystemGroup"
		title="Danger zone"
		subtitle="Delete this user group"
		:padding="true"
	>
		<p class="inline-notification error">
			Deleting {{ group.name }} removes the group and its role assignments. Users are not deleted.
		</p>
		<div role="toolbar" class="danger-zone-actions">
			<button type="button" class="bad" @click="confirmDeleteOpen = true">Delete user group</button>
		</div>
	</Section>
</template>

<style scoped>
.group-info {
	display: grid;
	grid-template-columns: 200px 1fr;
	column-gap: 1em;
	margin: 0;
}

.group-info dt {
	font-weight: bold;
}

.group-info dd {
	margin: 0;
}

.tag {
	font-size: 0.75rem;
	margin-left: 0.35rem;
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
