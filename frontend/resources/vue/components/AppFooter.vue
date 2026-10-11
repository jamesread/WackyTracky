<template>
	<footer title="footer">
		<p>
			<img
				v-if="logoUrl"
				class="logo"
				:src="logoUrl"
				:alt="appName + ' logo'"
				title="application icon"
				style="height: 1em;"
			>
			{{ appName }}
		</p>

		<p>
			<span v-if="showVersionNumber && currentVersion">{{ currentVersion }}</span>
			<span v-if="docsUrl">
				<a :href="docsUrl" target="_blank" rel="noopener noreferrer">Docs</a>
			</span>
			<span v-if="issuesUrl">
				<a :href="issuesUrl" target="_blank" rel="noopener noreferrer">Raise an issue</a>
			</span>
			<slot name="extra-links" />
		</p>

		<p v-if="showUpdateLink">
			<a
				id="available-version"
				:href="updateUrl || '#'"
				target="_blank"
				rel="noopener noreferrer"
			>
				{{ availableVersion }}
			</a>
		</p>
	</footer>
</template>

<script setup>
import { computed } from 'vue'
import { showUpdateLinkFromInit } from '../../../js/modules/appInit.js'

const props = defineProps({
	appName: { type: String, required: true },
	logoUrl: { type: String, default: '' },
	currentVersion: { type: String, default: '' },
	showVersionNumber: { type: Boolean, default: true },
	showNewVersions: { type: Boolean, default: true },
	availableVersion: { type: String, default: '' },
	docsUrl: { type: String, default: '' },
	issuesUrl: { type: String, default: '' },
	updateUrl: { type: String, default: '' },
})

const showUpdateLink = computed(() =>
	showUpdateLinkFromInit({
		showVersionNumber: props.showVersionNumber,
		showNewVersions: props.showNewVersions,
		availableVersion: props.availableVersion,
	}),
)
</script>
