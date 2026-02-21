<template>
	<Header
		ref="headerRef"
		title="WackyTracky"
		:logo-url="logoUrl"
		:sidebar-enabled="false"
		:breadcrumbs="false"
	>
		<template #toolbar>
			<div class="toolbar-row">
				<input
					ref="taskInputRef"
					:value="toolbarInputValue"
					type="text"
					class="toolbar-input"
					:placeholder="toolbarPlaceholder"
					:disabled="!!editingTaskId"
					@input="onToolbarInput"
					@keydown.enter="onToolbarKeydownEnter"
					@keydown.escape="onToolbarKeydownEscape"
				/>
				<button
					v-if="!searchMode && !editingTaskId"
					type="button"
					class="toolbar-submit-btn"
					:disabled="submittingTask"
					@click="submitTask"
				>
					<span v-if="submittingTask" class="toolbar-submit-spinner" aria-hidden="true"></span>
					<span v-else>Add</span>
				</button>
			</div>
			<div v-if="searchMode" class="toolbar-saved-searches">
				<button
					v-for="s in savedSearches"
					:key="s.id"
					type="button"
					class="saved-search-btn"
					@click="runSavedSearch(s.query)"
					@contextmenu.prevent="onSavedSearchContextMenu(s)"
				>
					{{ s.name }}
				</button>
				<button type="button" class="toolbar-submit-btn" @click="saveCurrentSearch">
					Save this search
				</button>
			</div>
		</template>
		<template #user-info>
			<div class="user-info">
				<button type="button" class="options-btn display-mode-btn" :title="displayModeLabel" aria-label="Display mode" @click="cycleDisplayMode">
					<HugeiconsIcon :icon="displayModeIcon" width="1.1em" height="1.1em" />
					<span class="display-mode-label">{{ displayModeLabel }}</span>
				</button>
				<button type="button" class="options-btn repo-status-btn" @click="showRepoStatus" title="Git status of todotxt directory" aria-label="Repo status">
					<HugeiconsIcon :icon="GitBranchIcon" width="1.1em" height="1.1em" />
				</button>
				<button type="button" class="options-btn options-settings-btn" @click="goToOptions" title="Options" aria-label="Options">
					<HugeiconsIcon :icon="Settings01Icon" width="1.1em" height="1.1em" />
				</button>
			</div>
		</template>
	</Header>
	<div v-if="repoStatusModal" class="repo-status-overlay" @click.self="repoStatusModal = null">
		<div class="repo-status-dialog" role="dialog" aria-labelledby="repo-status-title" aria-modal="true">
			<h2 id="repo-status-title" class="repo-status-title">Repo status (todotxt directory)</h2>
			<pre class="repo-status-output">{{ repoStatusOutput }}</pre>
			<div class="repo-status-actions">
				<button type="button" class="repo-status-close" @click="repoStatusModal = null">Close</button>
			</div>
		</div>
	</div>
	<div v-if="toastMessage" class="toast" :class="'toast-' + toastType" role="status" aria-live="polite">
		{{ toastMessage }}
	</div>
	<div v-if="saveSearchModalOpen" class="save-search-overlay" role="dialog" aria-labelledby="save-search-title" aria-modal="true" @click.self="saveSearchModalOpen = false">
		<div class="save-search-dialog">
			<h2 id="save-search-title" class="save-search-title">Save saved search</h2>
			<form @submit.prevent="submitSaveSearch" class="save-search-form">
				<label for="save-search-name">Name for this saved search</label>
				<input
					id="save-search-name"
					v-model="saveSearchNameDraft"
					type="text"
					class="save-search-input"
					autocomplete="off"
				/>
				<p v-if="saveSearchError" class="save-search-error" role="alert">{{ saveSearchError }}</p>
				<div class="save-search-actions">
					<button type="button" class="save-search-btn" @click="saveSearchModalOpen = false">Cancel</button>
					<button type="submit" class="save-search-btn save-search-btn-primary">Save</button>
				</div>
			</form>
		</div>
	</div>
	<div v-if="deleteSearchConfirm" class="delete-search-overlay" role="dialog" aria-labelledby="delete-search-title" aria-modal="true" @click.self="deleteSearchConfirm = null">
		<div class="delete-search-dialog">
			<h2 id="delete-search-title" class="delete-search-title">Delete saved search</h2>
			<p class="delete-search-message">Delete saved search "{{ deleteSearchConfirm?.name ?? 'this search' }}"?</p>
			<div class="delete-search-actions">
				<button type="button" class="delete-search-btn" @click="deleteSearchConfirm = null">Cancel</button>
				<button type="button" class="delete-search-btn delete-search-btn-danger" @click="confirmDeleteSearch">Delete</button>
			</div>
		</div>
	</div>
	<div v-if="shortcutsModal" class="shortcuts-overlay" role="dialog" aria-labelledby="shortcuts-title" aria-modal="true" @click.self="shortcutsModal = false">
		<div class="shortcuts-dialog">
			<h2 id="shortcuts-title" class="shortcuts-title">Keyboard shortcuts</h2>
			<div class="shortcuts-sections">
				<section class="shortcuts-section">
					<h3 class="shortcuts-section-title">Search</h3>
					<dl class="shortcuts-list">
						<dt><kbd>Ctrl</kbd> + <kbd>F</kbd></dt>
						<dd>Focus search</dd>
					</dl>
				</section>
				<section class="shortcuts-section">
					<h3 class="shortcuts-section-title">General</h3>
					<dl class="shortcuts-list">
						<dt><kbd>Ctrl</kbd> + <kbd>N</kbd></dt>
						<dd>Focus add task input</dd>
						<dt><kbd>m</kbd></dt>
						<dd>Cycle display mode (Hierarchy / Next actions / Waiting)</dd>
						<dt><kbd>Escape</kbd></dt>
						<dd>Cancel edit or exit search</dd>
					</dl>
				</section>
				<section class="shortcuts-section">
					<h3 class="shortcuts-section-title">Navigation</h3>
					<dl class="shortcuts-list">
						<dt><kbd>j</kbd> or <kbd>↓</kbd></dt>
						<dd>Move down</dd>
						<dt><kbd>k</kbd> or <kbd>↑</kbd></dt>
						<dd>Move up</dd>
						<dt><kbd>Escape</kbd></dt>
						<dd>Clear row focus</dd>
					</dl>
				</section>
				<section class="shortcuts-section">
					<h3 class="shortcuts-section-title">Edit</h3>
					<dl class="shortcuts-list">
						<dt><kbd>Enter</kbd> or <kbd>e</kbd> or <kbd>F2</kbd></dt>
						<dd>Start inline edit (when row focused)</dd>
						<dt><kbd>Escape</kbd></dt>
						<dd>Cancel edit / close dialogs</dd>
					</dl>
				</section>
				<section class="shortcuts-section">
					<h3 class="shortcuts-section-title">Actions</h3>
					<dl class="shortcuts-list">
						<dt><kbd>d</kbd> <kbd>d</kbd> or <kbd>Delete</kbd></dt>
						<dd>Mark task done</dd>
						<dt><kbd>y</kbd> <kbd>y</kbd></dt>
						<dd>Yank (copy) task</dd>
						<dt><kbd>p</kbd></dt>
						<dd>Paste (create task from yanked content)</dd>
						<dt><kbd>w</kbd> or <kbd>W</kbd></dt>
						<dd>Set wait until</dd>
						<dt><kbd>u</kbd> or <kbd>U</kbd></dt>
						<dd>Set due date</dd>
						<dt><kbd>i</kbd></dt>
						<dd>Add subtask (focus add task with current row as parent)</dd>
						<dt><kbd>r</kbd></dt>
						<dd>Randomize list order (current view only, not persisted)</dd>
					</dl>
				</section>
			</div>
			<div class="shortcuts-actions">
				<p class="shortcuts-hint">Press <kbd>?</kbd> to open this dialog.</p>
				<button type="button" class="shortcuts-close-btn" ref="shortcutsCloseRef" @click="shortcutsModal = false">Close</button>
			</div>
		</div>
	</div>
	<Navigation ref="navigation">
		<div id="layout">
			<div id="content">
				<main>
					<router-view :key="$route.fullPath" />
				</main>
				<footer class="app-footer">
					<span>WackyTracky</span>
					<button type="button" class="footer-shortcuts-btn" @click="openShortcutsDialog" aria-label="Keyboard shortcuts">
						Shortcuts
					</button>
				</footer>
			</div>
		</div>
	</Navigation>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick, computed, provide } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { HugeiconsIcon } from '@hugeicons/vue';
import Header from 'picocrank/vue/components/Header.vue';
import Navigation from 'picocrank/vue/components/Navigation.vue';
import { GitBranchIcon, PinIcon, Settings01Icon, Folder01Icon, PlayIcon, AlarmClockIcon } from '@hugeicons/core-free-icons';

const router = useRouter();
const route = useRoute();
const navigation = ref(null);
const headerRef = ref(null);
const taskInputRef = ref(null);
const taskDraft = ref('');
const editingTaskId = ref(null);
const editingDraft = ref('');
const searchMode = ref(false);
const searchInputValue = ref('');
const refreshTrigger = ref(0);
let searchDebounce = null;
const toastMessage = ref('');
const toastType = ref('success'); // 'success' | 'error'
let toastTimeout = null;
function showToast(message, type = 'success') {
	toastMessage.value = message;
	toastType.value = type;
	if (toastTimeout) clearTimeout(toastTimeout);
	toastTimeout = setTimeout(() => {
		toastMessage.value = '';
		toastTimeout = null;
	}, type === 'error' ? 6000 : 3000);
}
function toastErrorReason(e) {
	return e?.message || String(e);
}
const submittingTask = ref(false);
const repoStatusModal = ref(false);
const repoStatusOutput = ref('');
const shortcutsModal = ref(false);
const shortcutsCloseRef = ref(null);
const saveSearchModalOpen = ref(false);
const saveSearchNameDraft = ref('');
const saveSearchQuery = ref('');
const saveSearchError = ref('');
const deleteSearchConfirm = ref(null);
const pendingParentTaskId = ref(null);
const DISPLAY_MODE_KEY = 'wackytracky_display_mode';
const DISPLAY_MODES = ['hierarchy', 'onlyNextAction', 'onlyWaiting'];
function loadDisplayMode() {
	try {
		const stored = localStorage.getItem(DISPLAY_MODE_KEY);
		if (stored && DISPLAY_MODES.includes(stored)) return stored;
	} catch {}
	return 'hierarchy';
}
const displayMode = ref(loadDisplayMode());
watch(displayMode, (v) => {
	try {
		localStorage.setItem(DISPLAY_MODE_KEY, v);
	} catch {}
});
function cycleDisplayMode() {
	const i = DISPLAY_MODES.indexOf(displayMode.value);
	displayMode.value = DISPLAY_MODES[(i + 1) % DISPLAY_MODES.length];
}
const displayModeLabel = computed(() => {
	if (displayMode.value === 'hierarchy') return 'Hierarchy';
	if (displayMode.value === 'onlyNextAction') return 'Next actions';
	if (displayMode.value === 'onlyWaiting') return 'Waiting';
	return displayMode.value;
});
const displayModeIcon = computed(() => {
	if (displayMode.value === 'hierarchy') return Folder01Icon;
	if (displayMode.value === 'onlyNextAction') return PlayIcon;
	if (displayMode.value === 'onlyWaiting') return AlarmClockIcon;
	return Folder01Icon;
});
const focusFirstListItemTrigger = ref(0);

const SAVED_SEARCHES_KEY = 'wackytracky_saved_searches';
const savedSearches = ref([]);

async function loadSavedSearches() {
	try {
		if (window.client) {
			const res = await window.client.getSavedSearches({});
			const list = res.savedSearches || [];
			savedSearches.value = list.map((s) => ({ id: s.id, name: s.name, query: s.query }));
			return;
		}
	} catch {
		// fallback to localStorage
	}
	try {
		const raw = localStorage.getItem(SAVED_SEARCHES_KEY);
		savedSearches.value = raw ? JSON.parse(raw) : [];
	} catch {
		savedSearches.value = [];
	}
}

function persistSavedSearches() {
	localStorage.setItem(SAVED_SEARCHES_KEY, JSON.stringify(savedSearches.value));
	if (window.client) {
		window.client
			.setSavedSearches({
				savedSearches: savedSearches.value.map((s) => ({ id: s.id, name: s.name, query: s.query })),
			})
			.catch((e) => {
				showToast('Could not sync saved searches: ' + (e?.message || String(e)), 'error');
			});
	}
}

function addSavedSearch(name, query) {
	const q = (query || '').trim();
	if (!q) return;
	const id = `saved-${Date.now()}-${Math.random().toString(36).slice(2)}`;
	savedSearches.value = [...savedSearches.value, { id, name: (name || q).trim() || q, query: q }];
	persistSavedSearches();
}

function removeSavedSearch(id) {
	savedSearches.value = savedSearches.value.filter((s) => s.id !== id);
	persistSavedSearches();
}
function runSavedSearch(query) {
	const q = (query || '').trim();
	if (q) {
		searchInputValue.value = q;
		router.push({ path: '/search', query: { q } });
	}
}
function saveCurrentSearch() {
	const q = (route.query.q || searchInputValue.value || '').trim();
	if (!q) return;
	saveSearchQuery.value = q;
	saveSearchNameDraft.value = q;
	saveSearchError.value = '';
	saveSearchModalOpen.value = true;
}

function submitSaveSearch() {
	const name = (saveSearchNameDraft.value ?? '').trim();
	if (!name) {
		saveSearchError.value = 'Enter a name.';
		return;
	}
	const q = saveSearchQuery.value || '';
	if (!q) {
		saveSearchModalOpen.value = false;
		return;
	}
	addSavedSearch(name, q);
	showToast('Saved search added');
	saveSearchModalOpen.value = false;
	saveSearchNameDraft.value = '';
	saveSearchQuery.value = '';
	saveSearchError.value = '';
}

function onSavedSearchContextMenu(s) {
	deleteSearchConfirm.value = s;
}

function confirmDeleteSearch() {
	const s = deleteSearchConfirm.value;
	if (!s) return;
	removeSavedSearch(s.id);
	showToast('Saved search removed');
	deleteSearchConfirm.value = null;
}

const logoUrl = new URL('../images/logos/wacky-tracky.png', import.meta.url).href;

const toolbarPlaceholder = computed(() => {
	if (searchMode.value) return 'Search tasks...';
	return editingTaskId.value ? 'Editing task below…' : 'Add a task...';
});

const toolbarInputValue = computed(() => {
	if (searchMode.value) return searchInputValue.value;
	if (editingTaskId.value) return '';
	return taskDraft.value;
});

function startEdit(task) {
	editingTaskId.value = task.id;
	const base = (task.content || '').trim();
	const tagPart = (task.tags || []).length
		? ' ' + (task.tags || []).map((t) => '#' + t).join(' ')
		: '';
	const contextPart = (task.contexts || []).length
		? ' ' + (task.contexts || []).map((c) => '@' + c).join(' ')
		: '';
	editingDraft.value = base + tagPart + contextPart;
}

function cancelEdit() {
	editingTaskId.value = null;
	editingDraft.value = '';
}

async function submitEdit() {
	const content = (editingDraft.value || '').trim();
	if (!editingTaskId.value || !content || !window.client) {
		cancelEdit();
		return;
	}
	try {
		await window.client.updateTask({ id: editingTaskId.value, content });
		cancelEdit();
		refreshTrigger.value++;
		showToast('Task updated');
	} catch (e) {
		showToast('Could not save: ' + toastErrorReason(e), 'error');
		// keep editing and draft
	}
}

function onToolbarInput(e) {
	const v = e.target.value;
	if (searchMode.value) {
		searchInputValue.value = v;
		if (searchDebounce) clearTimeout(searchDebounce);
		searchDebounce = setTimeout(() => {
			const q = (searchInputValue.value || '').trim();
			if (q) router.replace({ path: '/search', query: { q } });
			else router.replace('/');
		}, 300);
	} else if (!editingTaskId.value) {
		taskDraft.value = v;
	}
}

function onToolbarKeydownEnter(e) {
	if (searchMode.value) {
		e.preventDefault();
		const q = (searchInputValue.value || '').trim();
		if (q) router.replace({ path: '/search', query: { q } });
	} else {
		submitTask();
	}
}

function onToolbarKeydownEscape(e) {
	if (searchMode.value) {
		e.preventDefault();
		searchMode.value = false;
		searchInputValue.value = '';
		const listId = route.params.listId;
		router.replace(listId ? '/lists/' + listId : '/');
	} else {
		e.preventDefault();
		cancelEdit();
		taskInputRef.value?.blur();
		focusFirstListItemTrigger.value++;
	}
}

function onCtrlF(e) {
	if (e.ctrlKey && e.key === 'f') {
		e.preventDefault();
		searchMode.value = true;
		searchInputValue.value = route.query.q || '';
		nextTick(() => taskInputRef.value?.focus());
	}
}

function onCtrlN(e) {
	if (e.ctrlKey && e.key === 'n') {
		e.preventDefault();
		if (!searchMode.value && !editingTaskId.value) {
			focusAddTaskInput(null);
		}
	}
}

function onKeyM(e) {
	if (e.key === 'm' && !e.ctrlKey && !e.metaKey && !e.altKey) {
		const tag = document.activeElement?.tagName?.toLowerCase();
		if (tag === 'input' || tag === 'textarea' || document.activeElement?.isContentEditable) return;
		e.preventDefault();
		cycleDisplayMode();
	}
}

async function submitTask() {
	const content = (taskDraft.value || '').trim();
	if (!content || !window.client) return;
	submittingTask.value = true;
	try {
		await window.client.createTask({
			content,
			parentListId: route.params.listId || '',
			parentTaskId: pendingParentTaskId.value || '',
		});
		taskDraft.value = '';
		pendingParentTaskId.value = null;
		refreshTrigger.value++;
		showToast('Task added');
	} catch (e) {
		showToast('Could not save: ' + toastErrorReason(e), 'error');
		// keep draft
	} finally {
		submittingTask.value = false;
	}
}

function focusAddTaskInput(parentTaskId) {
	pendingParentTaskId.value = parentTaskId ?? null;
	cancelEdit();
	nextTick(() => taskInputRef.value?.focus());
}

function setSearchQuery(term) {
	searchMode.value = true;
	searchInputValue.value = term;
	const q = (term || '').trim();
	if (q) router.push({ path: '/search', query: { q } });
	nextTick(() => taskInputRef.value?.focus());
}
provide('startEdit', startEdit);
provide('editingTaskId', editingTaskId);
provide('focusAddTaskInput', focusAddTaskInput);
provide('editingDraft', editingDraft);
provide('cancelEdit', cancelEdit);
provide('submitEdit', submitEdit);
provide('refreshTrigger', refreshTrigger);
provide('focusFirstListItemTrigger', focusFirstListItemTrigger);
provide('setSearchQuery', setSearchQuery);
provide('displayMode', displayMode);
provide('savedSearches', savedSearches);
provide('addSavedSearch', addSavedSearch);
provide('removeSavedSearch', removeSavedSearch);
provide('showToast', showToast);
provide('openShortcutsDialog', openShortcutsDialog);

async function getLists() {
	if (!navigation.value) return;
	navigation.value.clearNavigationLinks();
	navigation.value.addRouterLink('Welcome');
	navigation.value.addSeparator('nav-lists');
	if (!window.client) return;
	const ret = await window.client.getLists();
	for (const list of ret.lists) {
		navigation.value.addNavigationLink({
			name: list.id,
			title: list.title,
			path: `/lists/${list.id}`,
			to: `/lists/${list.id}`,
			icon: PinIcon,
		});
	}
}

function goToOptions() {
	router.push('/options');
}

function openShortcutsDialog() {
	shortcutsModal.value = true;
	nextTick(() => {
		shortcutsCloseRef.value?.focus();
	});
}

function onShortcutsKeydown(e) {
	if (e.key === 'Escape') {
		shortcutsModal.value = false;
		e.preventDefault();
	}
}
function onShortcutsOpenKeydown(e) {
	const tag = document.activeElement?.tagName?.toLowerCase();
	if (tag === 'input' || tag === 'textarea' || document.activeElement?.isContentEditable) return;
	if (e.key === '?' || (e.shiftKey && e.key === 'F1')) {
		e.preventDefault();
		openShortcutsDialog();
	}
}

async function showRepoStatus() {
	repoStatusOutput.value = 'Loading…';
	repoStatusModal.value = true;
	try {
		const res = await window.client.repoStatus({});
		repoStatusOutput.value = (res.output || '').trim() || '(no output)';
	} catch (e) {
		repoStatusOutput.value = 'Error: ' + (e?.message || String(e));
	}
}

function goHome() {
	router.push('/');
}

watch(
	() => route.params.listId,
	(listId) => {
		if (!listId) {
			editingTaskId.value = null;
			editingDraft.value = '';
			taskDraft.value = '';
		}
	}
);
watch(
	() => [route.params.listId, route.query.q],
	([listId, q]) => {
		if (listId) searchMode.value = false;
		else {
			searchMode.value = true;
			searchInputValue.value = q || '';
		}
	},
	{ immediate: true }
);

let headerLogoClickCleanup = null;

watch(refreshTrigger, () => { getLists(); });
watch(shortcutsModal, (open) => {
	if (open) {
		document.addEventListener('keydown', onShortcutsKeydown);
	} else {
		document.removeEventListener('keydown', onShortcutsKeydown);
	}
});
onMounted(() => {
	loadSavedSearches();
	getLists();
	document.addEventListener('keydown', onCtrlF);
	document.addEventListener('keydown', onCtrlN);
	document.addEventListener('keydown', onKeyM);
	document.addEventListener('keydown', onShortcutsOpenKeydown);
	// Clicking header logo/title navigates to index
	nextTick(() => {
		const el = headerRef.value?.$el;
		const branding = el?.querySelector?.('.image-and-title');
		if (branding) {
			branding.style.cursor = 'pointer';
			branding.addEventListener('click', goHome);
			headerLogoClickCleanup = () => {
				branding.removeEventListener('click', goHome);
			};
		}
	});
});
onUnmounted(() => {
	document.removeEventListener('keydown', onCtrlF);
	document.removeEventListener('keydown', onCtrlN);
	document.removeEventListener('keydown', onKeyM);
	document.removeEventListener('keydown', onShortcutsOpenKeydown);
	headerLogoClickCleanup?.();
});
</script>

<style scoped>
.toolbar-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex: 1;
	min-width: 0;
}
.toolbar-input {
	flex: 1;
	min-width: 0;
	max-width: 24rem;
	padding: 0.35rem 0.6rem;
	border-radius: 4px;
	border: 1px solid rgba(255, 255, 255, 0.3);
	background: rgba(255, 255, 255, 0.15);
	color: #fff;
	font-size: 0.9rem;
}
.toolbar-input::placeholder {
	color: rgba(255, 255, 255, 0.6);
}
.toolbar-input:focus {
	outline: none;
	border-color: rgba(255, 255, 255, 0.6);
	background: rgba(255, 255, 255, 0.2);
}
.toolbar-submit-btn {
	padding: 0.35rem 0.6rem;
	border-radius: 4px;
	border: 1px solid rgba(255, 255, 255, 0.5);
	background: transparent;
	color: #fff;
	font-size: 0.875rem;
	cursor: pointer;
}
.toolbar-submit-btn:hover {
	background: rgba(255, 255, 255, 0.15);
}
.toolbar-saved-searches {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: 0.35rem;
}
.saved-search-btn {
	padding: 0.25rem 0.5rem;
	border: 1px solid rgba(255, 255, 255, 0.5);
	border-radius: 4px;
	background: rgba(255, 255, 255, 0.1);
	color: #fff;
	font-size: 0.8rem;
	cursor: pointer;
}
.saved-search-btn:hover {
	background: rgba(255, 255, 255, 0.2);
}
.save-search-btn {
	padding: 0.25rem 0.5rem;
	border: 1px solid rgba(255, 255, 255, 0.5);
	border-radius: 4px;
	background: transparent;
	color: rgba(255, 255, 255, 0.9);
	font-size: 0.8rem;
	cursor: pointer;
}
.save-search-btn:hover {
	background: rgba(255, 255, 255, 0.15);
}
.options-settings-btn {
	margin-right: 0.5rem;
}
.user-info {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}
.options-btn {
	display: inline-flex;
	align-items: center;
	gap: 0.35rem;
	padding: 0.35rem 0.6rem;
	border-radius: 4px;
	border: 1px solid rgba(255, 255, 255, 0.5);
	background: transparent;
	color: #fff;
	font-size: 0.875rem;
	cursor: pointer;
}
.options-btn:hover {
	background: rgba(255, 255, 255, 0.15);
}
.display-mode-label {
	margin-left: 0.35rem;
	font-size: 0.85em;
}
.repo-status-overlay {
	position: fixed;
	inset: 0;
	background: rgba(0, 0, 0, 0.4);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1000;
}
.repo-status-dialog {
	background: #fff;
	border-radius: 0.5rem;
	box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
	padding: 1.25rem;
	min-width: 24rem;
	max-width: 90vw;
	max-height: 80vh;
	display: flex;
	flex-direction: column;
}
.repo-status-title {
	margin: 0 0 0.75rem 0;
	font-size: 1.1rem;
}
.repo-status-output {
	flex: 1;
	min-height: 8rem;
	margin: 0 0 1rem 0;
	padding: 0.75rem;
	background: #f5f5f5;
	border-radius: 0.35rem;
	font-size: 0.85rem;
	overflow: auto;
	white-space: pre-wrap;
	word-break: break-all;
}
.repo-status-actions {
	display: flex;
	justify-content: flex-end;
}
.repo-status-close {
	padding: 0.4rem 0.75rem;
	border-radius: 0.35rem;
	border: 1px solid #ccc;
	font-size: 0.9rem;
	cursor: pointer;
	background: #fff;
}
.repo-status-close:hover {
	background: #f0f0f0;
}
.toast {
	position: fixed;
	bottom: 1.5rem;
	left: 50%;
	transform: translateX(-50%);
	padding: 0.5rem 1rem;
	border-radius: 0.35rem;
	font-size: 0.9rem;
	box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
	z-index: 1001;
	max-width: 90vw;
}
.toast-success {
	background: #0a7c42;
	color: #fff;
}
.toast-error {
	background: #c00;
	color: #fff;
}
.toolbar-submit-btn:disabled {
	opacity: 0.7;
	cursor: not-allowed;
}
.toolbar-submit-spinner {
	display: inline-block;
	width: 1em;
	height: 1em;
	border: 2px solid rgba(255, 255, 255, 0.4);
	border-top-color: #fff;
	border-radius: 50%;
	animation: toolbar-spin 0.7s linear infinite;
}
@keyframes toolbar-spin {
	to { transform: rotate(360deg); }
}

.app-footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.75rem;
	padding: 0.5rem 1rem;
	font-size: 0.875rem;
	color: #666;
}
.footer-shortcuts-btn {
	padding: 0.25rem 0.5rem;
	border: 1px solid #ccc;
	border-radius: 0.35rem;
	background: #fff;
	color: #555;
	font-size: 0.85rem;
	cursor: pointer;
}
.footer-shortcuts-btn:hover {
	background: #f5f5f5;
	color: #333;
}

.save-search-overlay,
.delete-search-overlay {
	position: fixed;
	inset: 0;
	background: rgba(0, 0, 0, 0.4);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1000;
}
.save-search-dialog,
.delete-search-dialog {
	background: #fff;
	border-radius: 0.5rem;
	box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
	padding: 1.25rem;
	min-width: 18rem;
	max-width: 90vw;
}
.save-search-title,
.delete-search-title {
	margin: 0 0 1rem 0;
	font-size: 1.1rem;
}
.save-search-form label {
	display: block;
	margin-bottom: 0.35rem;
	font-size: 0.95rem;
}
.save-search-input {
	display: block;
	width: 100%;
	box-sizing: border-box;
	padding: 0.4rem 0.6rem;
	margin-bottom: 0.5rem;
	border: 1px solid #ccc;
	border-radius: 0.35rem;
	font-size: 0.95rem;
}
.save-search-error {
	margin: 0 0 0.5rem 0;
	font-size: 0.875rem;
	color: #c00;
}
.save-search-actions,
.delete-search-actions {
	display: flex;
	justify-content: flex-end;
	gap: 0.5rem;
	margin-top: 1rem;
}
.save-search-btn,
.delete-search-btn {
	padding: 0.4rem 0.75rem;
	border: 1px solid #ccc;
	border-radius: 0.35rem;
	background: #fff;
	font-size: 0.95rem;
	cursor: pointer;
}
.save-search-btn:hover,
.delete-search-btn:hover {
	background: #f0f0f0;
}
.save-search-btn-primary {
	background: #333;
	color: #fff;
	border-color: #333;
}
.save-search-btn-primary:hover {
	background: #555;
}
.delete-search-message {
	margin: 0;
	font-size: 0.95rem;
	color: #333;
}
.delete-search-btn-danger {
	background: #c00;
	color: #fff;
	border-color: #c00;
}
.delete-search-btn-danger:hover {
	background: #a00;
}

.shortcuts-overlay {
	position: fixed;
	inset: 0;
	background: rgba(0, 0, 0, 0.4);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1000;
}
.shortcuts-dialog {
	background: #fff;
	border-radius: 0.5rem;
	box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
	padding: 1.25rem;
	min-width: 22rem;
	max-width: 90vw;
	max-height: 85vh;
	display: flex;
	flex-direction: column;
}
.shortcuts-title {
	margin: 0 0 1rem 0;
	font-size: 1.1rem;
}
.shortcuts-sections {
	overflow-y: auto;
	flex: 1;
	margin-bottom: 1rem;
}
.shortcuts-section {
	margin-bottom: 1rem;
}
.shortcuts-section:last-child {
	margin-bottom: 0;
}
.shortcuts-section-title {
	margin: 0 0 0.35rem 0;
	font-size: 0.95rem;
	font-weight: 600;
	color: #333;
}
.shortcuts-list {
	display: grid;
	grid-template-columns: auto 1fr;
	gap: 0.25rem 1.25rem;
	margin: 0;
	font-size: 0.9rem;
	align-items: baseline;
}
.shortcuts-list dt {
	margin: 0;
	font-weight: 500;
	color: #555;
}
.shortcuts-list kbd {
	display: inline-block;
	padding: 0.15rem 0.4rem;
	font-family: inherit;
	font-size: 0.85em;
	background: #f0f0f0;
	border: 1px solid #ccc;
	border-radius: 0.25rem;
	box-shadow: 0 1px 0 #ccc;
}
.shortcuts-list dd {
	margin: 0;
	color: #333;
}
.shortcuts-actions {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;
}
.shortcuts-hint {
	margin: 0;
	font-size: 0.875rem;
	color: #666;
}
.shortcuts-close-btn {
	padding: 0.4rem 0.75rem;
	border-radius: 0.35rem;
	border: 1px solid #ccc;
	font-size: 0.9rem;
	cursor: pointer;
	background: #fff;
}
.shortcuts-close-btn:hover {
	background: #f0f0f0;
}
</style>
