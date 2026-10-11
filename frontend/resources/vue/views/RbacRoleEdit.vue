<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, Key01Icon } from '@hugeicons/core-free-icons'
import { useAppStatus } from '../../../js/modules/appStatus.js'
import { hasPermission } from '../../../js/modules/rbacAccess.js'

const SYSTEM_ROLES = new Set(['superuser', 'member'])

const route = useRoute()
const router = useRouter()
const status = useAppStatus()
const roleId = computed(() => Number(route.params.id))

const role = ref(null)
const permissions = ref([])
const selectedPermissionIds = ref([])
const groupsWithRole = ref([])
const loading = ref(true)
const error = ref('')
const saveError = ref('')
const saving = ref(false)
const deleting = ref(false)
const deleteError = ref('')
const confirmDeleteOpen = ref(false)
const description = ref('')

const canManage = computed(() => hasPermission(status.status, 'rbac.manage'))
const isSuperuserRole = computed(() => role.value?.name === 'superuser')
const isSystemRole = computed(() => SYSTEM_ROLES.has(role.value?.name || ''))
const canEditPermissions = computed(() => canManage.value && !isSuperuserRole.value)

async function load() {
	loading.value = true
	error.value = ''
	try {
		const [rolesRes, permsRes, groupsRes] = await Promise.all([
			window.client.listRbacRoles({}),
			window.client.listRbacPermissions({}),
			window.client.getRbacRoleGroups({ roleId: roleId.value }).catch(() => ({ groups: [] })),
		])
		role.value = (rolesRes.roles || []).find((r) => r.id === roleId.value) || null
		permissions.value = permsRes.permissions || []
		groupsWithRole.value = groupsRes.groups || []
		if (!role.value) {
			error.value = 'Role not found'
			return
		}
		description.value = role.value.description || ''
		selectedPermissionIds.value = [...(role.value.permissionIds || [])]
	} catch (e) {
		error.value = e instanceof Error ? e.message : 'Load failed'
		role.value = null
	} finally {
		loading.value = false
	}
}

function togglePermission(id, checked) {
	const set = new Set(selectedPermissionIds.value)
	if (checked) set.add(id)
	else set.delete(id)
	selectedPermissionIds.value = [...set]
}

async function save() {
	if (!role.value || !canEditPermissions.value) return
	saving.value = true
	saveError.value = ''
	try {
		await window.client.updateRbacRole({
			id: role.value.id,
			name: role.value.name,
			description: description.value,
			permissionIds: selectedPermissionIds.value,
		})
		await load()
	} catch (e) {
		saveError.value = e instanceof Error ? e.message : String(e)
	} finally {
		saving.value = false
	}
}

async function removeRole() {
	if (!role.value) return
	deleting.value = true
	deleteError.value = ''
	try {
		await window.client.deleteRbacRole({ id: role.value.id })
		confirmDeleteOpen.value = false
		router.push({ name: 'rbac-roles' })
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
		aria-labelledby="confirm-delete-role-title"
		@click.self="confirmDeleteOpen = false"
	>
		<div class="confirm-dialog">
			<h2 id="confirm-delete-role-title">Delete role</h2>
			<p>Delete role "{{ role?.name }}"?</p>
			<p v-if="deleteError" class="inline-notification error">{{ deleteError }}</p>
			<div class="confirm-actions">
				<button type="button" class="neutral" @click="confirmDeleteOpen = false">Cancel</button>
				<button type="button" class="bad" :disabled="deleting" @click="removeRole">
					{{ deleting ? 'Deleting…' : 'Delete' }}
				</button>
			</div>
		</div>
	</div>

	<Section :title="role?.name || 'Role'" :icon="Key01Icon" :padding="true">
		<template #toolbar>
			<RouterLink :to="{ name: 'rbac-roles' }" class="button inline-icon neutral">
				<HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
				<span>Roles</span>
			</RouterLink>
			<button type="button" class="inline-icon neutral" :disabled="loading" @click="load">
				Refresh
			</button>
		</template>

		<p v-if="loading" class="subtle">Loading…</p>
		<p v-if="error" class="inline-notification error">{{ error }}</p>

		<template v-else-if="role">
			<p>
				<span v-if="isSuperuserRole" class="tag">all access</span>
				<span v-if="isSystemRole" class="tag">system role</span>
			</p>
			<p>{{ role.description }}</p>
			<p v-if="groupsWithRole.length">
				Used by groups:
				{{ groupsWithRole.map((g) => g.name).join(', ') }}
			</p>
			<p v-else class="subtle">Not assigned to any groups yet.</p>
		</template>
	</Section>

	<Section
		v-if="role && !loading"
		title="Permissions"
		subtitle="Tick permissions granted by this role."
		:padding="true"
	>
		<p v-if="isSuperuserRole" class="inline-notification note">
			The superuser role grants all permissions and cannot be edited.
		</p>
		<FormLayout v-else @submit.prevent="save">
			<FormField v-if="canEditPermissions" label="Description" for="role-description">
				<input id="role-description" v-model="description" type="text" :disabled="saving" />
			</FormField>
			<table class="perm-table data-table">
				<thead>
					<tr>
						<th>Grant</th>
						<th>Permission</th>
						<th>Description</th>
					</tr>
				</thead>
				<tbody>
					<tr v-for="p in permissions" :key="p.id">
						<td>
							<input
								type="checkbox"
								:checked="selectedPermissionIds.includes(p.id)"
								:disabled="!canEditPermissions || saving"
								@change="togglePermission(p.id, $event.target.checked)"
							/>
						</td>
						<td><code>{{ p.name }}</code></td>
						<td>{{ p.description || '—' }}</td>
					</tr>
				</tbody>
			</table>
			<p v-if="saveError" class="inline-notification error">{{ saveError }}</p>
			<p v-if="!canManage" class="subtle">You can view permissions but not change them.</p>
			<template v-if="canEditPermissions" #actions>
				<button type="submit" class="good" :disabled="saving">
					{{ saving ? 'Saving…' : 'Save' }}
				</button>
			</template>
		</FormLayout>
	</Section>

	<Section
		v-if="role && canManage && !isSystemRole"
		title="Danger zone"
		subtitle="Delete this role"
		:padding="true"
	>
		<p class="inline-notification error">
			Deleting {{ role.name }} removes the role from all groups.
		</p>
		<div role="toolbar" class="danger-zone-actions">
			<button type="button" class="bad" @click="confirmDeleteOpen = true">Delete role</button>
		</div>
	</Section>
</template>

<style scoped>
.tag {
	font-size: 0.75rem;
	margin-right: 0.5rem;
}

.perm-table {
	width: 100%;
	margin: 1rem 0;
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
