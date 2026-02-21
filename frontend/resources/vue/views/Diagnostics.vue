<template>
	<section class="diagnostics-view">
		<h2 class="diagnostics-heading">Diagnostics</h2>

		<div class="diagnostics-section">
			<h3 class="diagnostics-section-title">Server diagnostics</h3>
			<button type="button" class="diagnostics-refresh-btn" @click="refreshServer">Refresh</button>
			<pre class="diagnostics-output" aria-label="Server diagnostics">{{ serverOutput }}</pre>
		</div>

		<div class="diagnostics-section">
			<h3 class="diagnostics-section-title">Browser diagnostics</h3>
			<button type="button" class="diagnostics-refresh-btn" @click="refreshBrowser">Refresh</button>
			<pre class="diagnostics-output" aria-label="Browser diagnostics">{{ browserOutput }}</pre>
		</div>
	</section>
</template>

<script setup>
import { ref } from 'vue';

const serverOutput = ref('');
const browserOutput = ref('');

function formatKv(lines) {
	return lines.filter(Boolean).join('\n');
}

async function refreshServer() {
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
</script>

<style scoped>
.diagnostics-view {
	padding: 1.25rem;
}

.diagnostics-heading {
	margin: 0 0 1.25rem 0;
	font-size: 1.25rem;
}

.diagnostics-section {
	margin-bottom: 2rem;
}

.diagnostics-section-title {
	margin: 0 0 0.5rem 0;
	font-size: 1.1rem;
	font-weight: 600;
}

.diagnostics-refresh-btn {
	margin-bottom: 0.5rem;
	padding: 0.4rem 0.75rem;
	font-size: 0.9rem;
	cursor: pointer;
}

.diagnostics-output {
	margin: 0;
	padding: 0.75rem;
	background: var(--femtocrank-bg-muted, #f5f5f5);
	border: 1px solid var(--border-color, #e1e5e9);
	border-radius: 0.35rem;
	font-size: 0.85rem;
	white-space: pre-wrap;
	word-break: break-all;
	min-height: 4rem;
}
</style>
