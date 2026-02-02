<template>
	<header>
		<div class = "logo-and-title" id = "sidebar-button">
			<img src = "../images/logos/wacky-tracky.png" alt = "wacky-tracky logo" class = "logo" />
			<h1 class = "fg1">wacky-tracky</h1>

			<button id = "sidebar-toggler-button" @click = "sidebar.toggle()">
				<HugeiconsIcon :icon="Menu07Icon" height = "1em" width = "1em" />
			</button>
		</div>

		<slot id = "slot-task-input" />
		<Breadcrumbs />
	</header>
	<div id = "layout">
		<Sidebar ref = "sidebar"/>
		<div id = "content">
			<main>
				<router-view :key = "$route.fullPath" />
			</main>
			<footer>
				<span>wacky-tracky</span>
			</footer>
		</div>
	</div>

</template>

<script setup>

import { ref } from 'vue';
import { onMounted } from 'vue';

import Sidebar from 'picocrank/vue/components/Sidebar.vue';
import Breadcrumbs from 'picocrank/vue/components/Breadcrumbs.vue';
import { HugeiconsIcon } from '@hugeicons/vue';
import { Menu07Icon } from '@hugeicons/core-free-icons';

const sidebar = ref(null);

async function getLists() {
	console.log(window.client);

	const ret = await window.client.getLists();

	for (const list of ret.lists) {
		sidebar.value.addNavigationLink({
			id: list.id,
			title: list.title,
			path: `/lists/${list.id}`,
		})
	}
}

onMounted(() => {
	sidebar.value.addNavigationLink({
		id: 'welcome',
		title: 'Welcome',
		path: '/',
		icon: 'mdi-home',
	});

	getLists()

	sidebar.value.open();
	sidebar.value.stick();
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
