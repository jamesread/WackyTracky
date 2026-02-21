<template>
	<Section title="Settings">
		<div class="options-section">
			<h3 class="options-subheading">Keyboard shortcuts</h3>
			<p class="options-hint">View and use keyboard shortcuts from the list view.</p>
			<button type="button" class="options-btn" @click="openShortcutsDialog">Open shortcuts</button>
		</div>
		<div class="options-section">
			<label for="options-datetime-format" class="options-label">Date/time display format</label>
			<select
				id="options-datetime-format"
				v-model="dateTimeDisplayFormat"
				class="options-select"
				aria-describedby="options-datetime-format-desc"
			>
				<option
					v-for="opt in dateTimeFormatOptions"
					:key="opt.value"
					:value="opt.value"
				>
					{{ opt.label }}
				</option>
			</select>
			<p id="options-datetime-format-desc" class="options-hint">
				How due dates and "wait until" times are shown in the task list and details.
			</p>
		</div>
	</Section>
</template>

<script setup>
import { inject } from 'vue';
import Section from 'picocrank/vue/components/Section.vue';
import { useSettings } from '../composables/useSettings.js';

const openShortcutsDialog = inject('openShortcutsDialog', () => {});
const {
	dateTimeDisplayFormat,
	dateTimeFormatOptions,
} = useSettings();
</script>

<style scoped>
.options-section {
	margin-bottom: 1.25rem;
}
.options-label {
	display: block;
	margin-bottom: 0.35rem;
	font-weight: 500;
	font-size: 0.95rem;
}
.options-select {
	padding: 0.4rem 0.6rem;
	min-width: 12rem;
	border: 1px solid #ccc;
	border-radius: 0.35rem;
	font-size: 0.95rem;
	background: var(--femtocrank-bg, #fff);
}
.options-hint {
	margin: 0.35rem 0 0 0;
	color: #666;
	font-size: 0.875rem;
}
.options-subheading {
	margin: 0 0 0.35rem 0;
	font-size: 1rem;
	font-weight: 600;
}
.options-btn {
	padding: 0.4rem 0.75rem;
	margin-top: 0.5rem;
	border: 1px solid #ccc;
	border-radius: 0.35rem;
	background: var(--femtocrank-bg, #fff);
	font-size: 0.95rem;
	cursor: pointer;
}
.options-btn:hover {
	background: #f0f0f0;
}
</style>
