<template>
	<Section
		title="Settings"
		:icon="Settings01Icon"
		subtitle="Server and client settings."
		:padding="true"
	>
		<FormLayout @submit.prevent>
			<FormField
				label="Keyboard shortcuts"
				description="View and use keyboard shortcuts from the list view."
			>
				<button type="button" class="neutral" @click="openShortcutsDialog">
					Open shortcuts
				</button>
			</FormField>

			<FormField
				label="Server name"
				for="options-server-name"
				description="Used in git commit messages when you sync from the header: wt sync: <server name>."
			>
				<input
					id="options-server-name"
					v-model="serverName"
					type="text"
					placeholder="e.g. homelab"
					autocomplete="off"
				/>
			</FormField>

			<FormField
				label="Git status"
				description="View git status output for the todo.txt data directory (requires network)."
			>
				<button
					type="button"
					class="neutral"
					:disabled="!isOnline || repoSyncRunning"
					@click="showRepoStatus"
				>
					Show git status
				</button>
			</FormField>

			<FormField
				label="Date/time display format"
				for="options-datetime-format"
				description="How due dates and wait-until times are shown in the task list and details."
			>
				<select id="options-datetime-format" v-model="dateTimeDisplayFormat">
					<option
						v-for="opt in dateTimeFormatOptions"
						:key="opt.value"
						:value="opt.value"
					>
						{{ opt.label }}
					</option>
				</select>
			</FormField>

			<FormField
				label="Use monospace font"
				description="Apply monospace font across the entire UI."
				component-has-label
			>
				<RadioGroup
					v-model="useMonospaceFont"
					name="options-monospace-font"
					variant="boolean"
					:options="booleanOptions"
				/>
			</FormField>

			<FormField
				label="Zen mode"
				description="Apply minimal styling for a distraction-free experience."
				component-has-label
			>
				<RadioGroup
					v-model="zenMode"
					name="options-zen-mode"
					variant="boolean"
					:options="booleanOptions"
				/>
			</FormField>

			<FormField
				label="Hide footer"
				description="Hide the footer element at the bottom of the page."
				component-has-label
			>
				<RadioGroup
					v-model="hideFooter"
					name="options-hide-footer"
					variant="boolean"
					:options="booleanOptions"
				/>
			</FormField>
		</FormLayout>
	</Section>
</template>

<script setup>
import { inject } from 'vue';
import FormField from 'picocrank/vue/components/FormField.vue';
import FormLayout from 'picocrank/vue/components/FormLayout.vue';
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue';
import Section from 'picocrank/vue/components/Section.vue';
import { Settings01Icon } from '@hugeicons/core-free-icons';
import { useSettings } from '../composables/useSettings.js';

const booleanOptions = [
	{ label: 'Yes', value: true },
	{ label: 'No', value: false },
];

const openShortcutsDialog = inject('openShortcutsDialog', () => {});
const showRepoStatus = inject('showRepoStatus', () => {});
const isOnline = inject('isOnline', { value: true });
const repoSyncRunning = inject('repoSyncRunning', { value: false });
const {
	dateTimeDisplayFormat,
	dateTimeFormatOptions,
	useMonospaceFont,
	zenMode,
	hideFooter,
	serverName,
} = useSettings();
</script>
