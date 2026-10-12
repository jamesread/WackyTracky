<template>
	<template v-if="canAccessControlPanel">
		<Section
			title="System Control Panel"
			:icon="DashboardSquareSettingIcon"
			subtitle="System and app settings"
		>
			<template #toolbar>
				<button type="button" class="neutral" :disabled="loading" @click="refresh">Refresh</button>
			</template>
			<Navigation ref="localNavigation">
				<NavigationGrid />
			</Navigation>
		</Section>

		<Section
			v-if="showQuickStatus"
			title="Quick status"
			:icon="DashboardSquareSettingIcon"
			subtitle="Operational details for administrators."
			:padding="true"
		>
			<FormLayout>
				<FormField
					label="Task database driver"
					description="Configured backend for lists and tasks (database.driver)."
				>
					<code>{{ databaseDriver || '—' }}</code>
				</FormField>
			</FormLayout>
		</Section>
	</template>
</template>

<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { DashboardSquareSettingIcon, UserShield01Icon } from '@hugeicons/core-free-icons'
import { fetchAppStatus, useAppStatus } from '../../../js/modules/appStatus.js'
import {
	canAccessControlPanelFromStatus,
	canAccessIamFromStatus,
	canViewSystemStatusFromStatus,
} from '../../../js/modules/rbacAccess.js'

const router = useRouter()
const status = useAppStatus()
const loading = ref(false)
const localNavigation = ref(null)
const canAccessControlPanel = computed(() => canAccessControlPanelFromStatus(status.status))
const showQuickStatus = computed(() => canViewSystemStatusFromStatus(status.status))
const databaseDriver = computed(() => String(status.status?.databaseDriver || ''))

async function refresh() {
	loading.value = true
	try {
		const st = await fetchAppStatus({ force: true })
		if (!canAccessControlPanelFromStatus(st)) {
			router.push('/')
			return
		}
		await nextTick()
		populateHubTiles()
	} finally {
		loading.value = false
	}
}

function populateHubTiles() {
	const nav = localNavigation.value
	if (!nav) return
	nav.clearNavigationLinks()
	nav.addRouterLink('Settings')
	nav.addRouterLink('TaskPropertyProperties', 'TPPs')
	nav.addRouterLink('Diagnostics')
	if (canAccessIamFromStatus(status.status)) {
		nav.addCallback('IAM', () => router.push({ name: 'iam' }), {
			icon: UserShield01Icon,
			name: 'iam',
			description: 'Users, groups, and roles',
		})
	}
}

onMounted(refresh)
</script>
