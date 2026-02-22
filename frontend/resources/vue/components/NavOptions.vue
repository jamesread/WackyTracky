<template>
	<section class="nav-options-page">
		<h2 id="nav-options-title" class="nav-options-title">Options</h2>
		<NavigationGrid
			filter-type="route"
			:filter="excludeListLinks"
		/>
	</section>
</template>

<script setup>
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue';

function excludeListLinks(link) {
	return !link.path?.startsWith('/lists/');
}
</script>

<style scoped>
.nav-options-page {
	padding: 1.25rem;
}

.nav-options-title {
	margin: 0 0 1rem 0;
	font-size: 1.25rem;
}
</style>
