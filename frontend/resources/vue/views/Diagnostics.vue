<template>
	<Section
		title="Diagnostics"
		:icon="Settings01Icon"
		subtitle="Debug and health information."
		:padding="true"
	>
		<template #toolbar>
			<router-link :to="{ name: 'controlPanel' }" class="button inline-icon neutral">
				<HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
				<span>Control Panel</span>
			</router-link>
			<button
				type="button"
				class="inline-icon neutral"
				aria-label="Refresh all"
				title="Refresh all"
				:disabled="serverLoading"
				@click="refreshAll"
			>
				<HugeiconsIcon
					:icon="Refresh01Icon"
					width="1em"
					height="1em"
					:strokeWidth="iconStrokeWidth"
					aria-hidden="true"
				/>
			</button>
		</template>

		<div class="diagnostics-stack">
			<ReadOnlyTextArea
				v-model="serverOutput"
				label="Server diagnostics"
				:rows="8"
				placeholder="No server data yet."
			>
				<template #actions>
					<button
						type="button"
						class="neutral"
						:disabled="serverLoading"
						@click="refreshServer"
					>
						{{ serverLoading ? 'Refreshing…' : 'Refresh' }}
					</button>
				</template>
			</ReadOnlyTextArea>

			<ReadOnlyTextArea
				v-model="browserOutput"
				label="Browser diagnostics"
				:rows="10"
				placeholder="No browser data yet."
			>
				<template #actions>
					<button type="button" class="neutral" @click="refreshBrowser">
						Refresh
					</button>
				</template>
			</ReadOnlyTextArea>
		</div>
	</Section>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import Section from 'picocrank/vue/components/Section.vue';
import ReadOnlyTextArea from 'picocrank/vue/components/ReadOnlyTextArea.vue';
import { HugeiconsIcon } from '@hugeicons/vue';
import { ArrowLeft01Icon, Refresh01Icon, Settings01Icon } from '@hugeicons/core-free-icons';

const iconStrokeWidth = 2.5;
const serverOutput = ref('');
const browserOutput = ref('');
const serverLoading = ref(false);

function formatKv(lines) {
	return lines.filter(Boolean).join('\n');
}

async function refreshServer() {
	serverLoading.value = true;
	serverOutput.value = 'Loading…';
	const lines = [];
	try {
		if (window.client?.version) {
			const res = await window.client.version({});
			if (res?.version != null) lines.push('version: ' + res.version);
			if (res?.commit != null) lines.push('commit: ' + res.commit);
			if (res?.date != null) lines.push('date: ' + res.date);
		}
		serverOutput.value = lines.length ? formatKv(lines) : 'No server data available.';
	} catch (e) {
		serverOutput.value = formatKv(['error: ' + (e?.message || String(e))]);
	} finally {
		serverLoading.value = false;
	}
}

function refreshBrowser() {
	const lines = [];
	const nav = typeof navigator !== 'undefined' ? navigator : {};
	const screen = typeof window !== 'undefined' && window.screen ? window.screen : {};
	lines.push('user_agent: ' + (nav.userAgent || ''));
	lines.push('language: ' + (nav.language || ''));
	lines.push('platform: ' + (nav.platform || ''));
	lines.push('screen_width: ' + (screen.width ?? ''));
	lines.push('screen_height: ' + (screen.height ?? ''));
	lines.push('window_inner_width: ' + (window?.innerWidth ?? ''));
	lines.push('window_inner_height: ' + (window?.innerHeight ?? ''));
	lines.push('location: ' + (window?.location?.href ?? ''));
	browserOutput.value = formatKv(lines);
}

async function refreshAll() {
	refreshBrowser();
	await refreshServer();
}

onMounted(refreshAll);
</script>

<style scoped>
.diagnostics-stack {
	display: flex;
	flex-direction: column;
	gap: 1.5rem;
}
</style>
