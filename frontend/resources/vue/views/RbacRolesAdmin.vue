<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import {
	ArrowLeft01Icon,
	Key01Icon,
	PlusSignIcon,
	Refresh01Icon,
} from '@hugeicons/core-free-icons'
import { useAppStatus } from '../../../js/modules/appStatus.js'
import { hasPermission } from '../../../js/modules/rbacAccess.js'

const iconStrokeWidth = 2.5
const SYSTEM_ROLES = new Set(['superuser', 'member'])

const roles = ref([])
const loading = ref(true)
const error = ref('')
const createName = ref('')
const createDescription = ref('')
const createError = ref('')
const creating = ref(false)
const createDialog = ref(null)
const createNameInput = ref(null)
const status = useAppStatus()
const router = useRouter()

const canManage = computed(() => hasPermission(status.status, 'rbac.manage'))

const tableHeaders = [
	{ key: 'name', label: 'Name', sortable: true },
	{ key: 'description', label: 'Description', sortable: true },
	{ key: 'permissionCount', label: 'Permissions', sortable: true, width: '8rem' },
	{ key: 'tags', label: 'Tags', sortable: false, width: '10rem' },
]

const rows = computed(() =>
	roles.value.map((r) => {
		const tags = []
		if (r.name === 'superuser') tags.push('all access')
		if (SYSTEM_ROLES.has(r.name)) tags.push('system role')
		return {
			id: r.id,
			name: r.name,
			description: r.description || '—',
			permissionCount: (r.permissionIds || r.permissions || []).length,
			tags: tags.join(', '),
		}
	}),
)

async function load() {
	loading.value = true
	error.value = ''
	try {
		const res = await window.client.listRbacRoles({})
		roles.value = res.roles || []
	} catch (e) {
		error.value = e instanceof Error ? e.message : 'Load failed'
	} finally {
		loading.value = false
	}
}

function resetCreateForm() {
	createName.value = ''
	createDescription.value = ''
	createError.value = ''
	creating.value = false
}

function openCreateDialog() {
	resetCreateForm()
	createDialog.value?.showModal()
	nextTick(() => createNameInput.value?.focus())
}

function closeCreateDialog() {
	createDialog.value?.close()
}

async function submitCreate() {
	const name = createName.value.trim()
	if (!name) {
		createError.value = 'Name is required.'
		return
	}
	creating.value = true
	createError.value = ''
	try {
		const res = await window.client.createRbacRole({
			name,
			description: createDescription.value.trim(),
			permissionIds: [],
		})
		closeCreateDialog()
		if (res.role?.id) {
			router.push({ name: 'rbac-role', params: { id: res.role.id } })
			return
		}
		await load()
	} catch (e) {
		createError.value = e instanceof Error ? e.message : String(e)
	} finally {
		creating.value = false
	}
}

onMounted(load)
</script>

<template>
	<dialog ref="createDialog" class="dialog" @close="resetCreateForm">
		<h2>Create role</h2>
		<p>Roles grant permissions to members of groups that have the role.</p>
		<FormLayout @submit.prevent="submitCreate">
			<FormField label="Name" for="create-role-name">
				<input
					id="create-role-name"
					ref="createNameInput"
					v-model="createName"
					type="text"
					autocomplete="off"
					required
					:disabled="creating"
				/>
			</FormField>
			<FormField label="Description" for="create-role-description">
				<input
					id="create-role-description"
					v-model="createDescription"
					type="text"
					autocomplete="off"
					:disabled="creating"
				/>
			</FormField>
			<p v-if="createError" class="inline-notification error">{{ createError }}</p>
			<template #actions>
				<button type="button" class="neutral" :disabled="creating" @click="closeCreateDialog">Cancel</button>
				<button type="submit" class="good" :disabled="creating || !createName.trim()">
					{{ creating ? 'Creating…' : 'Create' }}
				</button>
			</template>
		</FormLayout>
	</dialog>

	<Section
		subtitle="Assign permissions to roles, then attach roles to groups."
		classes="rbac-roles-list"
		:padding="false"
	>
		<template #title>
			<span class="section-title-with-icon">
				<HugeiconsIcon :icon="Key01Icon" width="22" height="22" aria-hidden="true" />
				Roles
			</span>
		</template>

		<template #toolbar>
			<RouterLink :to="{ name: 'iam' }" class="button inline-icon neutral">
				<HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
				<span>IAM</span>
			</RouterLink>
			<button
				type="button"
				class="inline-icon neutral"
				aria-label="Refresh"
				title="Refresh"
				:disabled="loading"
				@click="load"
			>
				<HugeiconsIcon
					:icon="Refresh01Icon"
					width="1em"
					height="1em"
					:strokeWidth="iconStrokeWidth"
					aria-hidden="true"
				/>
			</button>
			<button
				v-if="canManage"
				type="button"
				class="inline-icon good"
				aria-label="Create role"
				title="Create role"
				:disabled="loading"
				@click="openCreateDialog"
			>
				<HugeiconsIcon
					:icon="PlusSignIcon"
					width="1em"
					height="1em"
					:strokeWidth="iconStrokeWidth"
					aria-hidden="true"
				/>
			</button>
		</template>

		<p v-if="error" class="inline-notification error list-banner-pad">{{ error }}</p>
		<div v-if="loading && !roles.length" class="list-banner-pad muted">Loading…</div>

		<template v-else>
			<p v-if="!roles.length" class="inline-notification note list-banner-pad">No roles yet.</p>
			<Table
				v-else
				class="list-table-wrap"
				:data="rows"
				:headers="tableHeaders"
				:show-pagination="roles.length > 10"
			>
				<template #cell-name="{ value, row }">
					<RouterLink :to="{ name: 'rbac-role', params: { id: row.id } }" class="title-link">
						<strong>{{ value }}</strong>
					</RouterLink>
				</template>
				<template #cell-tags="{ value }">
					<span v-if="value" class="tag">{{ value }}</span>
				</template>
			</Table>
		</template>
	</Section>
</template>

<style scoped>
.section-title-with-icon {
	display: inline-flex;
	align-items: center;
	gap: 0.45em;
	vertical-align: middle;
}

.list-banner-pad {
	padding-left: 1em;
	padding-right: 1em;
}

.list-table-wrap {
	margin-top: 0.5rem;
	margin-bottom: 1.5rem;
}

.title-link {
	color: var(--pico-primary);
	text-decoration: none;
}

.title-link:hover {
	text-decoration: underline;
}

.tag {
	font-size: 0.75rem;
}
</style>
