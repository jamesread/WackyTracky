<template>
	<Section
		title="TPPs"
		:icon="Settings01Icon"
		subtitle="Set properties per tag (#name) or context (@name)."
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
				aria-label="Refresh"
				title="Refresh"
				@click="fetchAll"
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

		<p class="tpp-intro">
			Supported: <strong>css</strong> (arbitrary CSS for the chip style attribute),
			<strong>hide-at-times</strong>.
		</p>
		<p v-if="tppError" class="inline-notification error">{{ tppError }}</p>
	</Section>

	<Section
		title="Hide-at-times rule builder"
		subtitle="Tasks with a matching tag or context are hidden when the rule evaluates to true."
		:padding="true"
	>
		<p class="tpp-rule-hint">
			Use <strong>D</strong> (day: Mon–Sun), <strong>H</strong> (hour 0–23),
			<strong>M</strong> (minute 0–59). Example:
			<code>D == "Sat" || D == "Sun"</code>
		</p>
		<p class="tpp-rule-status" aria-live="polite">
			Current: D = {{ ruleStatus.D }}, H = {{ ruleStatus.H }}, M = {{ ruleStatus.M }}
		</p>

		<FormLayout @submit.prevent="testRule">
			<FormField
				label="Expression"
				for="tpp-rule-expr"
				description='Example: D == "Sat" || D == "Sun"'
			>
				<input
					id="tpp-rule-expr"
					v-model="ruleTestExpression"
					type="text"
					placeholder='e.g. D == "Sat" || D == "Sun"'
				/>
			</FormField>
			<template #actions>
				<button type="submit" class="neutral">Test</button>
			</template>
		</FormLayout>

		<div
			v-if="ruleTestResult !== null"
			class="tpp-rule-result"
			:class="{
				'tpp-rule-result-ok': ruleTestResult?.compiles && ruleTestResult?.evalError === undefined,
				'tpp-rule-result-err': ruleTestResult?.compileError || ruleTestResult?.evalError,
			}"
		>
			<template v-if="!ruleTestResult.compiles">
				<span class="tpp-rule-result-label">Compiles:</span> No. {{ ruleTestResult.compileError }}
			</template>
			<template v-else>
				<span class="tpp-rule-result-label">Compiles:</span> Yes.
				<template v-if="ruleTestResult.evalError">
					<span class="tpp-rule-result-label">Eval error:</span> {{ ruleTestResult.evalError }}
				</template>
				<template v-else>
					<span class="tpp-rule-result-label">Result:</span>
					{{ ruleTestResult.result ? 'true' : 'false' }}
				</template>
			</template>
		</div>
	</Section>

	<Section
		title="Tags"
		subtitle="Per-tag properties: css (chip inline style), hide-at-times (expression)."
		:padding="true"
	>
		<div class="tpp-table-wrap">
			<div class="tpp-table-header">
				<span class="tpp-col-name">Tag</span>
				<span class="tpp-col-css">css</span>
				<span class="tpp-col-hide">hide-at-times</span>
				<span class="tpp-col-action"></span>
			</div>
			<div v-for="(props, name) in tagEntries" :key="'tag-' + name" class="tpp-table-row">
				<span class="tpp-col-name">#{{ name }}</span>
				<input
					type="text"
					v-model="tagCss[name]"
					class="tpp-col-css"
					placeholder="e.g. background: #f0f0f0; color: #333"
					@change="saveTagProp(name, 'css', tagCss[name] || '')"
				/>
				<input
					type="text"
					v-model="tagHideAtTimes[name]"
					class="tpp-col-hide"
					placeholder='e.g. D == "Sat"'
					@change="saveTagProp(name, 'hide-at-times', tagHideAtTimes[name] || '')"
				/>
				<button type="button" class="bad small" title="Remove" @click="clearTagProp(name)">
					Remove
				</button>
			</div>
		</div>

		<FormLayout @submit.prevent="addTag">
			<FormField label="Tag name" for="tpp-new-tag-name">
				<input id="tpp-new-tag-name" v-model="newTagName" type="text" placeholder="Tag name" />
			</FormField>
			<FormField label="css" for="tpp-new-tag-css">
				<input
					id="tpp-new-tag-css"
					v-model="newTagCss"
					type="text"
					placeholder="e.g. background: #f0f0f0; color: #333"
				/>
			</FormField>
			<template #actions>
				<button type="submit" class="good" :disabled="!newTagName.trim() || !newTagCss.trim()">
					Add tag
				</button>
			</template>
		</FormLayout>
	</Section>

	<Section
		title="Contexts"
		subtitle="Per-context properties: css (chip inline style), hide-at-times (expression)."
		:padding="true"
	>
		<div class="tpp-table-wrap">
			<div class="tpp-table-header">
				<span class="tpp-col-name">Context</span>
				<span class="tpp-col-css">css</span>
				<span class="tpp-col-hide">hide-at-times</span>
				<span class="tpp-col-action"></span>
			</div>
			<div v-for="(props, name) in contextEntries" :key="'ctx-' + name" class="tpp-table-row">
				<span class="tpp-col-name">@{{ name }}</span>
				<input
					type="text"
					v-model="contextCss[name]"
					class="tpp-col-css"
					placeholder="e.g. background: #f0f0f0; color: #333"
					@change="saveContextProp(name, 'css', contextCss[name] || '')"
				/>
				<input
					type="text"
					v-model="contextHideAtTimes[name]"
					class="tpp-col-hide"
					placeholder='e.g. D == "Sat"'
					@change="saveContextProp(name, 'hide-at-times', contextHideAtTimes[name] || '')"
				/>
				<button type="button" class="bad small" title="Remove" @click="clearContextProp(name)">
					Remove
				</button>
			</div>
		</div>

		<FormLayout @submit.prevent="addContext">
			<FormField label="Context name" for="tpp-new-context-name">
				<input
					id="tpp-new-context-name"
					v-model="newContextName"
					type="text"
					placeholder="Context name"
				/>
			</FormField>
			<FormField label="css" for="tpp-new-context-css">
				<input
					id="tpp-new-context-css"
					v-model="newContextCss"
					type="text"
					placeholder="e.g. background: #f0f0f0; color: #333"
				/>
			</FormField>
			<template #actions>
				<button
					type="submit"
					class="good"
					:disabled="!newContextName.trim() || !newContextCss.trim()"
				>
					Add context
				</button>
			</template>
		</FormLayout>
	</Section>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue';
import Section from 'picocrank/vue/components/Section.vue';
import FormField from 'picocrank/vue/components/FormField.vue';
import FormLayout from 'picocrank/vue/components/FormLayout.vue';
import { HugeiconsIcon } from '@hugeicons/vue';
import { ArrowLeft01Icon, Refresh01Icon, Settings01Icon } from '@hugeicons/core-free-icons';

const iconStrokeWidth = 2.5;
const refreshTaskPropertyProperties = inject('refreshTaskPropertyProperties', () => Promise.resolve());
const tagProperties = ref({});
const contextProperties = ref({});
const tagCss = ref({});
const contextCss = ref({});
const tagHideAtTimes = ref({});
const contextHideAtTimes = ref({});
const newTagName = ref('');
const newTagCss = ref('');
const newContextName = ref('');
const newContextCss = ref('');
const tppError = ref('');
const ruleStatus = ref({ D: '', H: 0, M: 0 });
const ruleTestExpression = ref('');
const ruleTestResult = ref(null);

const tagEntries = computed(() => tagProperties.value);
const contextEntries = computed(() => contextProperties.value);

async function fetchTpp() {
	tppError.value = '';
	if (!window.client) return;
	try {
		const res = await window.client.getTaskPropertyProperties({});
		tagProperties.value = res.tagProperties || {};
		contextProperties.value = res.contextProperties || {};
		const tagCssMap = {};
		const tagHides = {};
		for (const [name, obj] of Object.entries(tagProperties.value)) {
			tagCssMap[name] = obj?.props?.css ?? '';
			tagHides[name] = obj?.props?.['hide-at-times'] ?? '';
		}
		const ctxCssMap = {};
		const ctxHides = {};
		for (const [name, obj] of Object.entries(contextProperties.value)) {
			ctxCssMap[name] = obj?.props?.css ?? '';
			ctxHides[name] = obj?.props?.['hide-at-times'] ?? '';
		}
		tagCss.value = tagCssMap;
		contextCss.value = ctxCssMap;
		tagHideAtTimes.value = tagHides;
		contextHideAtTimes.value = ctxHides;
	} catch (e) {
		tppError.value = e?.message || String(e);
	}
}

async function setProp(propertyType, propertyName, key, value) {
	if (!window.client) return;
	try {
		await window.client.setTaskPropertyProperty({
			propertyType,
			propertyName: propertyName.trim(),
			key,
			value: (value || '').trim(),
		});
		await fetchTpp();
		await refreshTaskPropertyProperties();
	} catch (e) {
		tppError.value = e?.message || String(e);
	}
}

function saveTagProp(name, key, value) {
	setProp('tag', name, key, value);
}

function clearTagProp(name) {
	setProp('tag', name, 'css', '');
	setProp('tag', name, 'hide-at-times', '');
}

function saveContextProp(name, key, value) {
	setProp('context', name, key, value);
}

function clearContextProp(name) {
	setProp('context', name, 'css', '');
	setProp('context', name, 'hide-at-times', '');
}

async function fetchRuleStatus() {
	if (!window.client) return;
	try {
		const res = await window.client.ruleStatus({});
		ruleStatus.value = { D: res.D ?? '', H: res.H ?? 0, M: res.M ?? 0 };
	} catch {
		// ignore
	}
}

async function testRule() {
	ruleTestResult.value = null;
	if (!window.client) return;
	const expr = (ruleTestExpression.value || '').trim();
	if (!expr) return;
	try {
		const res = await window.client.ruleTest({ expression: expr });
		ruleTestResult.value = {
			compiles: res.compiles,
			compileError: res.compileError || undefined,
			result: res.result,
			evalError: res.evalError || undefined,
		};
	} catch (e) {
		ruleTestResult.value = { compiles: false, compileError: e?.message || String(e) };
	}
}

async function addTag() {
	const name = (newTagName.value || '').trim();
	const css = (newTagCss.value || '').trim();
	if (!name || !css) return;
	await setProp('tag', name, 'css', css);
	newTagName.value = '';
	newTagCss.value = '';
}

async function addContext() {
	const name = (newContextName.value || '').trim();
	const css = (newContextCss.value || '').trim();
	if (!name || !css) return;
	await setProp('context', name, 'css', css);
	newContextName.value = '';
	newContextCss.value = '';
}

async function fetchAll() {
	await fetchTpp();
	await fetchRuleStatus();
}

onMounted(fetchAll);
</script>

<style scoped>
.tpp-intro,
.tpp-rule-hint {
	margin: 0 0 1rem;
	color: var(--muted-text-color, #666);
	font-size: 0.9rem;
}

.tpp-rule-hint code {
	font-size: 0.85em;
	background: var(--standout-bg-color, #f5f5f5);
	padding: 0.1rem 0.3rem;
	border-radius: 0.2rem;
}

.tpp-rule-status {
	margin: 0 0 1rem;
	font-size: 0.9rem;
}

.tpp-table-wrap {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
	margin-bottom: 1.25rem;
}

.tpp-table-header,
.tpp-table-row {
	display: grid;
	grid-template-columns: minmax(5rem, auto) 10rem 1fr auto;
	gap: 0.5rem;
	align-items: center;
}

.tpp-table-header {
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--muted-text-color, #666);
	padding-bottom: 0.25rem;
	border-bottom: 1px solid var(--border-color, #ccc);
}

.tpp-col-name {
	font-weight: 500;
}

.tpp-col-css,
.tpp-col-hide {
	width: 100%;
	min-width: 0;
}

.tpp-rule-result {
	margin-top: 1rem;
	font-size: 0.875rem;
}

.tpp-rule-result-ok {
	color: var(--muted-text-color, #666);
}

.tpp-rule-result-err {
	color: var(--karma-bad, #c00);
}

.tpp-rule-result-label {
	font-weight: 500;
	margin-right: 0.25rem;
}
</style>
