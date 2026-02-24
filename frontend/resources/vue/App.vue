<template>
	<div v-if="!isOnline" class="offline-banner" role="status" aria-live="polite">
		You are offline. Editing and adding tasks (inbox only) are saved locally and will sync when you reconnect.
	</div>
	<Header
		ref="headerRef"
		title="WackyTracky"
		:logo-url="logoUrl"
		:sidebar-enabled="false"
		:breadcrumbs="false"
		:show-branding="!isPwa"
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
			</div>
			<div v-if="searchMode && isOnline" ref="savedSearchSelectRef" class="toolbar-saved-searches toolbar-saved-searches-select">
				<div class="saved-search-select-trigger" tabindex="0" :class="{ open: savedSearchDropdownOpen }" @click="savedSearchDropdownOpen = !savedSearchDropdownOpen" @keydown.enter.prevent="savedSearchDropdownOpen = !savedSearchDropdownOpen" @keydown.space.prevent="savedSearchDropdownOpen = !savedSearchDropdownOpen" @keydown.down.prevent="openSavedSearchDropdownAndFocusFilter" role="combobox" :aria-expanded="savedSearchDropdownOpen" aria-haspopup="listbox" :aria-label="selectedSavedSearch ? selectedSavedSearch.name : 'Saved searches'">
					<span class="saved-search-select-value">{{ selectedSavedSearch ? selectedSavedSearch.name : 'Saved searches…' }}</span>
				</div>
				<div v-if="savedSearchDropdownOpen" class="saved-search-select-dropdown" role="listbox">
					<input
						ref="savedSearchFilterInputRef"
						v-model="savedSearchFilter"
						type="text"
						class="saved-search-select-filter"
						placeholder="Filter saved searches…"
						@keydown.escape="savedSearchDropdownOpen = false"
						@keydown.stop
					/>
					<ul class="saved-search-select-list">
						<li
							v-for="s in filteredSavedSearches"
							:key="s.id"
							role="option"
							:aria-selected="currentSearchQuery === s.query"
							class="saved-search-select-option"
							@click="selectSavedSearch(s)"
							@contextmenu.prevent="onSavedSearchContextMenu(s)"
						>
							{{ s.name }}
						</li>
						<li v-if="filteredSavedSearches.length === 0" class="saved-search-select-empty">No saved searches match</li>
					</ul>
					<button type="button" class="saved-search-select-save-btn" @click="saveCurrentSearchFromSelect">
						Save this search
					</button>
				</div>
			</div>
		</template>
		<template #user-info>
			<div class="user-info">
				<button type="button" class="options-btn display-mode-btn" :title="displayModeLabel" aria-label="Display mode" @click="cycleDisplayMode">
					<HugeiconsIcon :icon="displayModeIcon" width="1.1em" height="1.1em" />
					<span class="display-mode-label">{{ displayModeLabel }}</span>
				</button>
				<button type="button" class="options-btn repo-status-btn" :disabled="!isOnline" @click="showRepoStatus" title="Git status of todotxt directory (requires network)" aria-label="Repo status">
					<HugeiconsIcon :icon="GitBranchIcon" width="1.1em" height="1.1em" />
				</button>
				<button type="button" class="options-btn options-settings-btn" @click="goToOptions" title="Options" aria-label="Options">
					<HugeiconsIcon :icon="Settings01Icon" width="1.1em" height="1.1em" />
				</button>
			</div>
		</template>
	</Header>
	<div v-if="navOptionsOpen" class="nav-options-overlay" @click.self="navOptionsOpen = false">
		<NavOptions />
	</div>
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
						<dt><kbd>Ctrl</kbd> + <kbd>K</kbd></dt>
						<dd>Focus toolbar input, or cycle mode (add task ↔ search) when already focused</dd>
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
					<span>
						<a href="https://jamesread.github.io/WackyTracky/" target="_blank" rel="noopener noreferrer" class="footer-brand">WackyTracky</a>
					</span>
					<span>
						<button type="button" class="footer-shortcuts-btn" @click="openShortcutsDialog" aria-label="Keyboard shortcuts">
							Shortcuts
						</button>
					</span>
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
import { useOffline } from './composables/useOffline.js';
import { useSettings } from './composables/useSettings.js';
import {
	addOfflineTask,
	setOfflineEdit,
	removeOfflineTask,
	removeOfflineEdit,
	getOfflineTasks,
	getOfflineEdits,
	setCachedInbox,
	getCachedInbox,
	generateOfflineTaskId,
	INBOX_LIST_ID,
} from '../../js/modules/offlineStorage.js';

const router = useRouter();
const route = useRoute();
const { isOnline } = useOffline();
const { useMonospaceFont, zenMode, hideFooter } = useSettings();
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
const navOptionsOpen = ref(false);
const shortcutsModal = ref(false);
const shortcutsCloseRef = ref(null);
const saveSearchModalOpen = ref(false);
const saveSearchNameDraft = ref('');
const saveSearchQuery = ref('');
const saveSearchError = ref('');
const deleteSearchConfirm = ref(null);
const pendingParentTaskId = ref(null);
const isPwa = ref(false);
let displayModeMqStandalone = null;
let displayModeMqWco = null;
function updatePwaDisplayMode() {
	isPwa.value = displayModeMqStandalone?.matches || displayModeMqWco?.matches || false;
}
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
const savedSearchDropdownOpen = ref(false);
const savedSearchFilter = ref('');
const savedSearchSelectRef = ref(null);
const savedSearchFilterInputRef = ref(null);

const taskPropertyProperties = ref({ tagProperties: {}, contextProperties: {} });
async function loadTaskPropertyProperties() {
	try {
		if (window.client) {
			const res = await window.client.getTaskPropertyProperties({});
			taskPropertyProperties.value = {
				tagProperties: res.tagProperties || {},
				contextProperties: res.contextProperties || {},
			};
		}
	} catch {
		taskPropertyProperties.value = { tagProperties: {}, contextProperties: {} };
	}
}

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

const currentSearchQuery = computed(() => (route.query.q || '').trim());
const selectedSavedSearch = computed(() => {
	const q = currentSearchQuery.value;
	if (!q) return null;
	return savedSearches.value.find((s) => s.query === q) || null;
});
const filteredSavedSearches = computed(() => {
	const f = (savedSearchFilter.value || '').trim().toLowerCase();
	if (!f) return savedSearches.value;
	return savedSearches.value.filter(
		(s) => s.name.toLowerCase().includes(f) || (s.query || '').toLowerCase().includes(f)
	);
});

function selectSavedSearch(s) {
	runSavedSearch(s.query);
	savedSearchDropdownOpen.value = false;
	savedSearchFilter.value = '';
}

function saveCurrentSearchFromSelect() {
	saveCurrentSearch();
	savedSearchDropdownOpen.value = false;
}

function openSavedSearchDropdownAndFocusFilter() {
	savedSearchDropdownOpen.value = true;
	nextTick(() => savedSearchFilterInputRef.value?.focus());
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
	if (!editingTaskId.value || !content) {
		cancelEdit();
		return;
	}
	if (!isOnline.value) {
		setOfflineEdit(editingTaskId.value, content);
		cancelEdit();
		refreshTrigger.value++;
		showToast('Saved locally (will sync when online)');
		return;
	}
	if (!window.client) {
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

function onCtrlK(e) {
	if (!(e.ctrlKey || e.metaKey) || e.key !== 'k') return;
	e.preventDefault();
	if (editingTaskId.value) return;
	const inputFocused = document.activeElement === taskInputRef.value;
	if (inputFocused) {
		if (searchMode.value) {
			searchMode.value = false;
			searchInputValue.value = '';
			if (route.path === '/search') router.replace('/');
		} else {
			searchMode.value = true;
			searchInputValue.value = route.query.q || '';
			if (isOnline.value) {
				const q = (searchInputValue.value || '').trim();
				if (q) router.replace({ path: '/search', query: { q } });
			}
		}
	} else {
		nextTick(() => taskInputRef.value?.focus());
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
	if (!content) return;
	const listId = route.params.listId || '';
	const parentTaskId = pendingParentTaskId.value || '';
	if (!isOnline.value) {
		const cached = getCachedInbox();
		const inboxId = cached?.listId ?? INBOX_LIST_ID;
		if (listId !== inboxId) {
			showToast('Offline: you can only add tasks to Inbox.', 'error');
			return;
		}
		const id = generateOfflineTaskId();
		addOfflineTask({
			id,
			content,
			parentId: parentTaskId || INBOX_LIST_ID,
			parentType: parentTaskId ? 'task' : 'list',
			tags: [],
			contexts: [],
			createdAt: new Date().toISOString(),
		});
		taskDraft.value = '';
		pendingParentTaskId.value = null;
		refreshTrigger.value++;
		showToast('Saved locally (will sync when online)');
		return;
	}
	if (!window.client) return;
	submittingTask.value = true;
	try {
		await window.client.createTask({
			content,
			parentListId: listId,
			parentTaskId,
		});
		taskDraft.value = '';
		pendingParentTaskId.value = null;
		refreshTrigger.value++;
		showToast('Task added');
	} catch (e) {
		showToast('Could not save: ' + toastErrorReason(e), 'error');
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
provide('taskPropertyProperties', taskPropertyProperties);
provide('refreshTaskPropertyProperties', loadTaskPropertyProperties);
provide('isOnline', isOnline);
provide('cacheInbox', cacheList);

function cacheList(listId, listTitle, data) {
	setCachedInbox({ listId, listTitle, ...data });
}

async function syncOfflineChanges() {
	const tasks = getOfflineTasks();
	const edits = getOfflineEdits();
	if (tasks.length === 0 && Object.keys(edits).length === 0) return;
	if (!window.client) return;
	let synced = 0;
	for (const t of tasks) {
		try {
			await window.client.createTask({
				content: t.content,
				parentListId: INBOX_LIST_ID,
				parentTaskId: t.parentType === 'task' ? t.parentId : '',
			});
			removeOfflineTask(t.id);
			synced++;
		} catch (e) {
			showToast('Could not sync task: ' + (e?.message || String(e)), 'error');
		}
	}
	for (const [taskId, { content }] of Object.entries(edits)) {
		try {
			await window.client.updateTask({ id: taskId, content });
			removeOfflineEdit(taskId);
			synced++;
		} catch (e) {
			showToast('Could not sync edit: ' + (e?.message || String(e)), 'error');
		}
	}
	if (synced > 0) {
		refreshTrigger.value++;
		showToast('Offline changes synced');
	}
}

async function getLists() {
	if (!navigation.value) return;
	navigation.value.clearNavigationLinks();
	navigation.value.addRouterLink('Welcome');
	navigation.value.addSeparator('nav-lists');
	if (!isOnline.value) {
		const cached = getCachedInbox();
		if (cached) {
			navigation.value.addNavigationLink({
				name: cached.listId,
				title: cached.listTitle,
				path: `/lists/${cached.listId}`,
				to: `/lists/${cached.listId}`,
				icon: PinIcon,
				type: 'route',
			});
		} else {
			navigation.value.addNavigationLink({
				name: INBOX_LIST_ID,
				title: 'Inbox',
				path: `/lists/${INBOX_LIST_ID}`,
				to: `/lists/${INBOX_LIST_ID}`,
				icon: PinIcon,
				type: 'route',
			});
		}
	} else if (window.client) {
		const ret = await window.client.getLists();
		for (const list of ret.lists) {
			navigation.value.addNavigationLink({
				name: list.id,
				title: list.title,
				path: `/lists/${list.id}`,
				to: `/lists/${list.id}`,
				icon: PinIcon,
				type: 'route',
			});
		}
	}
	navigation.value.addSeparator('nav-settings');
	navigation.value.addRouterLink('TaskPropertyProperties', 'TPPs');
	navigation.value.addRouterLink('Settings', 'Settings');
	navigation.value.addRouterLink('Diagnostics', 'Diagnostics');
}

function goToOptions() {
	navOptionsOpen.value = true;
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
watch(() => route.path, (path) => {
	if (path !== '/options') navOptionsOpen.value = false;
});
watch(isOnline, (online) => {
	if (!online) {
		searchMode.value = false;
		if (route.path === '/' || route.path === '/search') {
			const cached = getCachedInbox();
			const targetId = cached?.listId ?? INBOX_LIST_ID;
			router.replace('/lists/' + targetId);
		} else if (route.params.listId) {
			const cached = getCachedInbox();
			const inboxId = cached?.listId ?? INBOX_LIST_ID;
			if (route.params.listId !== inboxId) {
				router.replace('/lists/' + inboxId);
			}
		}
	}
});

let headerLogoClickCleanup = null;

watch(refreshTrigger, () => { getLists(); });
watch(savedSearchDropdownOpen, (open) => {
	if (open) {
		savedSearchFilter.value = '';
		nextTick(() => savedSearchFilterInputRef.value?.focus());
	}
});
watch(shortcutsModal, (open) => {
	if (open) {
		document.addEventListener('keydown', onShortcutsKeydown);
	} else {
		document.removeEventListener('keydown', onShortcutsKeydown);
	}
});
watch(useMonospaceFont, (enabled) => {
	if (enabled) {
		document.body.classList.add('monospace-font');
	} else {
		document.body.classList.remove('monospace-font');
	}
}, { immediate: true });
watch(zenMode, (enabled) => {
	if (enabled) {
		document.body.classList.add('zen-mode');
	} else {
		document.body.classList.remove('zen-mode');
	}
}, { immediate: true });
watch(hideFooter, (enabled) => {
	if (enabled) {
		document.body.classList.add('hide-footer');
	} else {
		document.body.classList.remove('hide-footer');
	}
}, { immediate: true });
onMounted(() => {
	displayModeMqStandalone = window.matchMedia('(display-mode: standalone)');
	displayModeMqWco = window.matchMedia('(display-mode: window-controls-overlay)');
	updatePwaDisplayMode();
	displayModeMqStandalone.addEventListener('change', updatePwaDisplayMode);
	displayModeMqWco.addEventListener('change', updatePwaDisplayMode);
	loadSavedSearches();
	loadTaskPropertyProperties();
	getLists();
	window.addEventListener('online', syncOfflineChanges);
	document.addEventListener('keydown', onCtrlF);
	document.addEventListener('keydown', onCtrlN);
	document.addEventListener('keydown', onCtrlK);
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
	document.addEventListener('click', onSavedSearchSelectClickOutside);
});
onUnmounted(() => {
	if (displayModeMqStandalone) displayModeMqStandalone.removeEventListener('change', updatePwaDisplayMode);
	if (displayModeMqWco) displayModeMqWco.removeEventListener('change', updatePwaDisplayMode);
	window.removeEventListener('online', syncOfflineChanges);
	document.removeEventListener('keydown', onCtrlF);
	document.removeEventListener('keydown', onCtrlN);
	document.removeEventListener('keydown', onCtrlK);
	document.removeEventListener('keydown', onKeyM);
	document.removeEventListener('keydown', onShortcutsOpenKeydown);
	document.removeEventListener('click', onSavedSearchSelectClickOutside);
	headerLogoClickCleanup?.();
});

function onSavedSearchSelectClickOutside(e) {
	if (!savedSearchDropdownOpen.value) return;
	const el = savedSearchSelectRef.value;
	if (el && !el.contains(e.target)) savedSearchDropdownOpen.value = false;
}
</script>

<style scoped>
.offline-banner {
	background: #f59e0b;
	color: #fff;
	text-align: center;
	padding: 0.4rem 1rem;
	font-size: 0.875rem;
}
.toolbar-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex-grow: 2;
	min-width: 0;
}

/* Shared height, padding, margin for header controls */
.toolbar-input,
.saved-search-select-trigger,
.options-btn {
	box-sizing: border-box;
	height: 2rem;
	padding: 0.35rem 0.6rem;
	margin: 0;
	line-height: 1.25;
	font-size: 0.875rem;
	border-radius: 4px;
	border-width: 1px;
	border-style: solid;
}

.toolbar-input {
	flex: 1;
	min-width: 0;
	border-color: rgba(255, 255, 255, 0.3);
	background: rgba(255, 255, 255, 0.15);
	color: #fff;
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
	height: 2rem;
	padding: 0 0.6rem;
	margin: 0;
	box-sizing: border-box;
	border-radius: 4px;
	border: 1px solid rgba(255, 255, 255, 0.5);
	background: transparent;
	color: #fff;
	font-size: 0.875rem;
	line-height: 1.25;
	cursor: pointer;
}
.toolbar-submit-btn:hover {
	background: rgba(255, 255, 255, 0.15);
}
.toolbar-saved-searches {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: 0.5rem;
}
.toolbar-saved-searches-select {
	position: relative;
}
.saved-search-select-trigger {
	display: flex;
	align-items: center;
	min-width: 8rem;
	border-color: rgba(255, 255, 255, 0.5);
	background: rgba(255, 255, 255, 0.1);
	color: #fff;
	cursor: pointer;
}
.saved-search-select-trigger:hover,
.saved-search-select-trigger.open {
	background: rgba(255, 255, 255, 0.2);
}
.saved-search-select-value {
	display: block;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}
.saved-search-select-dropdown {
	position: absolute;
	top: 100%;
	left: 0;
	margin-top: 0.25rem;
	min-width: 100%;
	max-width: 20rem;
	max-height: 16rem;
	display: flex;
	flex-direction: column;
	background: #fff;
	border-radius: 4px;
	box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
	z-index: 100;
}
.saved-search-select-filter {
	width: 100%;
	padding: 0.4rem 0.6rem;
	border: none;
	border-bottom: 1px solid #e0e0e0;
	border-radius: 4px 4px 0 0;
	font-size: 0.875rem;
	box-sizing: border-box;
}
.saved-search-select-filter:focus {
	outline: none;
	border-color: #666;
}
.saved-search-select-list {
	margin: 0;
	padding: 0.25rem 0;
	list-style: none;
	overflow-y: auto;
	flex: 1;
	min-height: 0;
}
.saved-search-select-option {
	padding: 0.35rem 0.6rem;
	font-size: 0.875rem;
	color: #333;
	cursor: pointer;
}
.saved-search-select-option:hover {
	background: #f0f0f0;
}
.saved-search-select-empty {
	padding: 0.5rem 0.6rem;
	font-size: 0.875rem;
	color: #666;
}
.saved-search-select-save-btn {
	margin: 0.25rem;
	padding: 0.35rem 0.6rem;
	border: 1px solid #ccc;
	border-radius: 4px;
	background: #f8f8f8;
	font-size: 0.8rem;
	cursor: pointer;
}
.saved-search-select-save-btn:hover {
	background: #eee;
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

@media (display-mode: window-controls-overlay) {
	.user-info {
		padding-right: calc(100vw - env(titlebar-area-x, 0px) - env(titlebar-area-width, 100vw));
	}
}
.options-btn {
	display: inline-flex;
	align-items: center;
	gap: 0.35rem;
	border-color: rgba(255, 255, 255, 0.5);
	background: transparent;
	color: #fff;
	cursor: pointer;
}
.options-btn:hover {
	background: rgba(255, 255, 255, 0.15);
}
.display-mode-label {
	margin-left: 0;
	font-size: inherit;
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

#layout {
	display: flex;
	flex-direction: column;
	min-height: 100vh;
}
#content {
	display: flex;
	flex-direction: column;
	flex: 1;
}
#content main {
	flex: 1;
}
.app-footer {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 0.75rem;
	padding: 0.5rem 1rem;
	font-size: 0.75rem;
	color: #666;
}
.app-footer .footer-brand {
	color: inherit;
	text-decoration: none;
}
.app-footer .footer-brand:hover {
	text-decoration: underline;
}
.footer-shortcuts-btn {
	padding: 0;
	border: none;
	background: none;
	color: inherit;
	font-size: inherit;
	font-weight: normal;
	cursor: pointer;
	text-decoration: none;
}
.footer-shortcuts-btn:hover {
	text-decoration: underline;
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
