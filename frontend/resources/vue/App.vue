<template>
	<Header
		username="Guest"
		title="wacky-tracky"
		:logo-url="logoUrl"
		:sidebar-enabled="true"
		:show-branding="true"
		:breadcrumbs="true"
		:top-bar-enabled="false"
		:navigation="navigation"
		@toggle-sidebar="toggleSidebar"
	>
		<template #toolbar>
			<div id="slot-task-input" />
		</template>
	</Header>

	<Navigation ref="navigation">
		<div id="layout">
			<Sidebar ref="sidebar" />

			<div id="content">
				<main>
					<router-view :key="$route.fullPath" />
				</main>
				<footer>
					<span>wacky-tracky</span>
				</footer>
			</div>
		</div>
	</Navigation>
</template>

<script setup>
import { ref, onMounted } from 'vue';

import Header from 'picocrank/vue/components/Header.vue';
import Navigation from 'picocrank/vue/components/Navigation.vue';
import Sidebar from 'picocrank/vue/components/Sidebar.vue';

const logoUrl = new URL('../images/logos/wacky-tracky.png', import.meta.url).href;
const sidebar = ref(null);
const navigation = ref(null);

function toggleSidebar() {
	if (sidebar.value) {
		sidebar.value.toggle();
	}
}

async function getLists() {
	const ret = await window.client.getLists();
	const nav = navigation.value;
	if (!nav?.addNavigationLink) return;

	for (const list of ret.lists) {
		nav.addNavigationLink({
			name: String(list.id),
			title: list.title,
			path: `/lists/${list.id}`,
			type: 'route',
		});
	}
}

onMounted(() => {
	// Add links on the Navigation component (not Sidebar).
	const nav = navigation.value;
	if (nav?.addRouterLink) {
		nav.addRouterLink('Welcome');
	}

	getLists();

	if (sidebar.value) {
		sidebar.value.open();
		sidebar.value.stick();
	}
});
</script>

<style scoped>
header button {
	color: #fff;
	border: 1px solid transparent;
}
button {
	color: inherit;
}
</style>
