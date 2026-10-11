<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, Key01Icon, PlusSignIcon, Refresh01Icon } from '@hugeicons/core-free-icons'

const iconStrokeWidth = 2.5

const keys = ref([])
const loading = ref(true)
const error = ref('')
const name = ref('')
const readOnly = ref(false)
const newSecret = ref('')
const createError = ref('')
const creating = ref(false)
const createDialog = ref(null)
const confirmDelete = ref(null)

const booleanOptions = [
	{ label: 'Read-only', value: true },
	{ label: 'Read/write', value: false },
]

const tableHeaders = [
	{ key: 'name', label: 'Name', sortable: true },
	{ key: 'access', label: 'Access', sortable: true, width: '8rem' },
	{ key: 'actions', label: 'Actions', sortable: false, width: '8rem' },
]

const rows = computed(() =>
	keys.value.map((k) => ({
		id: k.id,
		name: k.name,
		access: k.readOnly ? 'read-only' : 'read/write',
		actions: '',
	})),
)

async function load() {
	loading.value = true
	error.value = ''
	try {
		const res = await window.client.listApiKeys({})
		keys.value = res.keys || []
	} catch (e) {
		error.value = e instanceof Error ? e.message : 'Load failed'
	} finally {
		loading.value = false
	}
}

function openCreateDialog() {
	name.value = ''
	readOnly.value = false
	createError.value = ''
	newSecret.value = ''
	createDialog.value?.showModal()
	nextTick(() => document.getElementById('api-key-name')?.focus())
}

function closeCreateDialog() {
	createDialog.value?.close()
}

async function createKey() {
	creating.value = true
	createError.value = ''
	newSecret.value = ''
	try {
		const res = await window.client.createApiKey({
			name: name.value || 'API key',
			readOnly: readOnly.value,
		})
		newSecret.value = res.secret || ''
		name.value = ''
		await load()
	} catch (e) {
		createError.value = e instanceof Error ? e.message : String(e)
	} finally {
		creating.value = false
	}
}

async function removeKey() {
	if (!confirmDelete.value) return
	try {
		await window.client.deleteApiKey({ id: confirmDelete.value.id })
		confirmDelete.value = null
		await load()
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
		confirmDelete.value = null
	}
}

onMounted(load)
</script>

<template>
	<dialog ref="createDialog" class="dialog">
		<h2>Create API key</h2>
		<p>Bearer tokens for automation and MCP. Copy the secret when shown — it will not appear again.</p>
		<FormLayout @submit.prevent="createKey">
			<FormField label="Key name" for="api-key-name">
				<input id="api-key-name" v-model="name" type="text" placeholder="API key" :disabled="creating" />
			</FormField>
			<FormField label="Access" component-has-label>
				<RadioGroup v-model="readOnly" name="api-key-readonly" :options="booleanOptions" />
			</FormField>
			<p v-if="createError" class="inline-notification error">{{ createError }}</p>
			<p v-if="newSecret" class="inline-notification note">
				Copy this secret now: <code>{{ newSecret }}</code>
			</p>
			<template #actions>
				<button type="button" class="neutral" :disabled="creating" @click="closeCreateDialog">
					{{ newSecret ? 'Close' : 'Cancel' }}
				</button>
				<button v-if="!newSecret" type="submit" class="good" :disabled="creating">
					{{ creating ? 'Creating…' : 'Create key' }}
				</button>
			</template>
		</FormLayout>
	</dialog>

	<div
		v-if="confirmDelete"
		class="confirm-overlay"
		role="dialog"
		aria-modal="true"
		aria-labelledby="confirm-delete-key-title"
		@click.self="confirmDelete = null"
	>
		<div class="confirm-dialog">
			<h2 id="confirm-delete-key-title">Delete API key</h2>
			<p>Delete API key "{{ confirmDelete.name }}"?</p>
			<div class="confirm-actions">
				<button type="button" class="neutral" @click="confirmDelete = null">Cancel</button>
				<button type="button" class="bad" @click="removeKey">Delete</button>
			</div>
		</div>
	</div>

	<Section
		subtitle="Bearer tokens for automation and MCP."
		classes="api-keys-list"
		:padding="false"
	>
		<template #title>
			<span class="section-title-with-icon">
				<HugeiconsIcon :icon="Key01Icon" width="22" height="22" aria-hidden="true" />
				API keys
			</span>
		</template>
		<template #toolbar>
			<RouterLink :to="{ name: 'userControlPanel' }" class="button inline-icon neutral">
				<HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
				<span>Back</span>
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
				type="button"
				class="inline-icon good"
				aria-label="Create API key"
				title="Create API key"
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
		<div v-if="loading && !keys.length" class="list-banner-pad muted">Loading…</div>
		<template v-else>
			<p v-if="!keys.length" class="inline-notification note list-banner-pad">No API keys yet.</p>
			<Table
				v-else
				class="list-table-wrap"
				:data="rows"
				:headers="tableHeaders"
				:show-pagination="keys.length > 10"
			>
				<template #cell-name="{ value }">
					<strong>{{ value }}</strong>
				</template>
				<template #cell-actions="{ row }">
					<div class="actions-cell">
						<button type="button" class="bad small" @click="confirmDelete = row">Delete</button>
					</div>
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

.actions-cell {
	text-align: right;
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
