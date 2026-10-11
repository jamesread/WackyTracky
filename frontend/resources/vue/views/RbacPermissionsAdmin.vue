<script setup>
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import { onMounted, ref } from 'vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon } from '@hugeicons/core-free-icons'

const permissions = ref([])
const error = ref('')
const loading = ref(true)

onMounted(async () => {
	loading.value = true
	try {
		const res = await window.client.listRbacPermissions({})
		permissions.value = res.permissions || []
		error.value = ''
	} catch (e) {
		error.value = e instanceof Error ? e.message : 'Load failed'
	} finally {
		loading.value = false
	}
})
</script>

<template>
	<Section title="Permission catalog" subtitle="Read-only catalog of RBAC permissions." :padding="false">
		<template #toolbar>
			<router-link :to="{ name: 'iam' }" class="button inline-icon neutral">
				<HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
				<span>IAM</span>
			</router-link>
		</template>
		<p v-if="error" class="inline-notification error list-banner-pad">{{ error }}</p>
		<div v-if="loading && !permissions.length" class="list-banner-pad muted">Loading…</div>
		<Table
			v-else-if="permissions.length"
			class="list-table-wrap"
			:data="permissions.map((p) => ({
				name: p.name,
				description: p.description || '—',
				roles: (p.roleNames || []).join(', ') || '—',
			}))"
			:headers="[
				{ key: 'name', label: 'Name', sortable: true },
				{ key: 'description', label: 'Description', sortable: true },
				{ key: 'roles', label: 'Roles', sortable: false },
			]"
		/>
		<p v-else-if="!loading" class="inline-notification note list-banner-pad">No permissions in catalog.</p>
	</Section>
</template>

<style scoped>
.list-banner-pad {
	padding-left: 1em;
	padding-right: 1em;
}

.list-table-wrap {
	margin-top: 0.5rem;
	margin-bottom: 1.5rem;
}
</style>
