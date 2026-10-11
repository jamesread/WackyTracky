<template>
	<Section v-if="canAccessControlPanel" title="System Control Panel" subtitle="System and app settings">
		<template #toolbar>
			<button type="button" class="neutral" :disabled="loading" @click="refresh">Refresh</button>
		</template>
		<Navigation ref="localNavigation">
			<NavigationGrid />
		</Navigation>
	</Section>
</template>

<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import { UserShield01Icon } from '@hugeicons/core-free-icons'
import { fetchAppStatus, useAppStatus } from '../../../js/modules/appStatus.js'
import {
	canAccessControlPanelFromStatus,
	canAccessIamFromStatus,
} from '../../../js/modules/rbacAccess.js'

const router = useRouter()
const status = useAppStatus()
const loading = ref(false)
const localNavigation = ref(null)
const canAccessControlPanel = computed(() => canAccessControlPanelFromStatus(status.status))

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
	const st = status.status
	let tileCount = 0
	if (canAccessIamFromStatus(st)) {
		tileCount += 1
		nav.addCallback('IAM', () => router.push({ name: 'iam' }), {
			icon: UserShield01Icon,
			name: 'iam',
			description: 'Users, groups, and roles',
		})
	}
	if (tileCount === 0) {
		router.push('/')
	}
}

onMounted(refresh)
</script>
