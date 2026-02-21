<template>
	<section class="tpp-view">
		<h2 id="tpp-title" class="tpp-title">Tag & context properties</h2>
		<p class="tpp-intro">Set properties per tag (#name) or context (@name). Supported: <strong>css</strong> (arbitrary CSS for the chip style attribute), <strong>hide-at-times</strong>.</p>

		<div class="tpp-section tpp-rule-builder">
			<h3 class="tpp-subheading">Hide-at-times rule builder</h3>
			<p class="tpp-rule-hint">Tasks with a tag or context that has a hide-at-times rule are hidden from the list when the rule evaluates to true. Use <strong>D</strong> (day: Mon–Sun), <strong>H</strong> (hour 0–23), <strong>M</strong> (minute 0–59). Example: <code>D == "Sat" || D == "Sun"</code></p>
			<p class="tpp-rule-status" aria-live="polite">Current: D = {{ ruleStatus.D }}, H = {{ ruleStatus.H }}, M = {{ ruleStatus.M }}</p>
			<div class="tpp-rule-test">
				<input v-model="ruleTestExpression" type="text" class="tpp-rule-expr-input" placeholder='e.g. D == "Sat" || D == "Sun"' @keydown.enter.prevent="testRule" />
				<button type="button" class="tpp-add-btn" @click="testRule">Test</button>
			</div>
			<div v-if="ruleTestResult !== null" class="tpp-rule-result" :class="{ 'tpp-rule-result-ok': ruleTestResult?.compiles && ruleTestResult?.evalError === undefined, 'tpp-rule-result-err': ruleTestResult?.compileError || ruleTestResult?.evalError }">
				<template v-if="!ruleTestResult.compiles">
					<span class="tpp-rule-result-label">Compiles:</span> No. {{ ruleTestResult.compileError }}
				</template>
				<template v-else>
					<span class="tpp-rule-result-label">Compiles:</span> Yes.
					<template v-if="ruleTestResult.evalError">
						<span class="tpp-rule-result-label">Eval error:</span> {{ ruleTestResult.evalError }}
					</template>
					<template v-else>
						<span class="tpp-rule-result-label">Result:</span> {{ ruleTestResult.result ? 'true' : 'false' }}
					</template>
				</template>
			</div>
		</div>

		<div class="tpp-section">
			<h3 class="tpp-subheading">Tags</h3>
			<p class="tpp-prop-hint">Per-tag properties: <strong>css</strong> (chip inline style), <strong>hide-at-times</strong> (expression to hide tasks with this tag).</p>
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
					<button type="button" class="tpp-remove-btn" title="Remove" @click="clearTagProp(name)">Remove</button>
				</div>
			</div>
			<div class="tpp-add-row">
				<input v-model="newTagName" type="text" class="tpp-name-input" placeholder="Tag name" />
				<input v-model="newTagCss" type="text" class="tpp-col-css tpp-add-input" placeholder="css" />
				<button type="button" class="tpp-add-btn" @click="addTag">Add tag</button>
			</div>
		</div>

		<div class="tpp-section">
			<h3 class="tpp-subheading">Contexts</h3>
			<p class="tpp-prop-hint">Per-context properties: <strong>css</strong> (chip inline style), <strong>hide-at-times</strong> (expression to hide tasks with this context).</p>
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
					<button type="button" class="tpp-remove-btn" title="Remove" @click="clearContextProp(name)">Remove</button>
				</div>
			</div>
			<div class="tpp-add-row">
				<input v-model="newContextName" type="text" class="tpp-name-input" placeholder="Context name" />
				<input v-model="newContextCss" type="text" class="tpp-col-css tpp-add-input" placeholder="css" />
				<button type="button" class="tpp-add-btn" @click="addContext">Add context</button>
			</div>
		</div>

		<p v-if="tppError" class="tpp-error">{{ tppError }}</p>
	</section>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue';

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

onMounted(() => {
	fetchTpp();
	fetchRuleStatus();
});
</script>

<style scoped>
.tpp-view {
	padding: 1.5rem;
}
.tpp-title {
	margin: 0 0 0.5rem 0;
	font-size: 1.25rem;
}
.tpp-intro {
	margin: 0 0 1.25rem 0;
	color: var(--femtocrank-text-muted, #666);
	font-size: 0.9rem;
}
.tpp-section {
	margin-bottom: 1.5rem;
}
.tpp-subheading {
	margin: 0 0 0.5rem 0;
	font-size: 1rem;
}
.tpp-prop-hint {
	margin: 0 0 0.5rem 0;
	font-size: 0.875rem;
	color: var(--femtocrank-text-muted, #666);
}
.tpp-table-wrap {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
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
	color: var(--femtocrank-text-muted, #666);
	padding-bottom: 0.25rem;
	border-bottom: 1px solid var(--femtocrank-border, #ccc);
}
.tpp-table-row {
	min-height: 2rem;
}
.tpp-col-name {
	font-weight: 500;
}
.tpp-col-css {
	width: 100%;
	min-width: 0;
	padding: 0.25rem 0.5rem;
	border: 1px solid var(--femtocrank-border, #ccc);
	border-radius: 0.25rem;
	font-size: 0.875rem;
}
.tpp-col-hide {
	width: 100%;
	min-width: 0;
	padding: 0.25rem 0.5rem;
	border: 1px solid var(--femtocrank-border, #ccc);
	border-radius: 0.25rem;
	font-size: 0.875rem;
}
.tpp-name-input {
	width: 8rem;
	padding: 0.25rem 0.5rem;
	border: 1px solid var(--femtocrank-border, #ccc);
	border-radius: 0.25rem;
}
.tpp-add-input {
	width: 10rem;
}
.tpp-remove-btn,
.tpp-add-btn {
	padding: 0.25rem 0.5rem;
	font-size: 0.875rem;
	border: 1px solid var(--femtocrank-border, #ccc);
	border-radius: 0.25rem;
	background: var(--femtocrank-bg, #fff);
	cursor: pointer;
}
.tpp-add-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	margin-top: 0.75rem;
}
.tpp-error {
	color: var(--femtocrank-error, #c00);
	margin: 1rem 0 0 0;
}

.tpp-rule-builder {
	background: var(--femtocrank-bg-muted, #f5f5f5);
	border-radius: 0.5rem;
	padding: 1rem;
}
.tpp-rule-hint {
	margin: 0 0 0.5rem 0;
	font-size: 0.875rem;
	color: var(--femtocrank-text-muted, #666);
}
.tpp-rule-hint code {
	font-size: 0.85em;
	background: var(--femtocrank-bg, #fff);
	padding: 0.1rem 0.3rem;
	border-radius: 0.2rem;
}
.tpp-rule-status {
	margin: 0 0 0.5rem 0;
	font-size: 0.9rem;
}
.tpp-rule-test {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	margin-bottom: 0.5rem;
}
.tpp-rule-expr-input {
	flex: 1;
	max-width: 24rem;
	padding: 0.25rem 0.5rem;
	border: 1px solid var(--femtocrank-border, #ccc);
	border-radius: 0.25rem;
}
.tpp-rule-result {
	margin-top: 0.5rem;
	font-size: 0.875rem;
}
.tpp-rule-result-ok {
	color: var(--femtocrank-text-muted, #666);
}
.tpp-rule-result-err {
	color: var(--femtocrank-error, #c00);
}
.tpp-rule-result-label {
	font-weight: 500;
	margin-right: 0.25rem;
}

</style>
