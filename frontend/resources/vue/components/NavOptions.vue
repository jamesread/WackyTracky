<template>
	<Section
		title="Options"
		:icon="Settings01Icon"
		subtitle="Shortcuts to settings, task properties, and diagnostics."
		:padding="true"
	>
		<NavigationGrid
			filter-type="route"
			:filter="excludeListLinks"
		/>
	</Section>
</template>

<script setup>
import Section from 'picocrank/vue/components/Section.vue';
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue';
import { Settings01Icon } from '@hugeicons/core-free-icons';

function excludeListLinks(link) {
	return !link.path?.startsWith('/lists/');
}
</script>
