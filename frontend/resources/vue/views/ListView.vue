<template>
	<Section :title="sectionTitle">
		<template #toolbar>
			<button v-if="listId && !searchQuery" type="button" class="list-options-btn" @click="openListOptions">List options</button>
		</template>
		<div v-if="searchQuery" class="search-heading">
			<p>Search results for “{{ searchQuery }}”</p>
		</div>
		<div v-if="loading" class="list-loading" aria-live="polite">
			<span class="list-loading-spinner" aria-hidden="true"></span>
			<p>Loading…</p>
		</div>
		<div v-else-if="searchError" class="list-search-error" role="alert">
			<p class="list-search-error-title">Search failed. Please try again.</p>
			<p class="list-search-error-reason">{{ searchError }}</p>
		</div>
		<div v-else-if="!items.length" class="list-empty">
			<p>{{ searchQuery ? 'No matching tasks.' : 'This list is empty.' }}</p>
		</div>
		<div v-else class="task-list-container">
			<ul>
				<li
					v-for="(row, index) in items"
					:key="row.task.id"
					class="task-row"
					:class="{ 'task-row-editing': currentEditingId === row.task.id, 'task-row-focused': !currentEditingId && focusedIndex === index, 'task-row-subtask': displayMode === 'hierarchy' && row.depth > 0, 'task-row-project': displayMode === 'hierarchy' && row.task.countSubitems > 0, 'task-row-yanked': yankedTask && row.task.id === yankedTask.id, 'task-row-waiting': isFutureWait(row.task) }"
					:style="displayMode === 'hierarchy' && row.depth > 0 ? { marginLeft: row.depth * 2.25 + 'rem' } : {}"
					@dblclick="currentEditingId !== row.task.id && injectedStartEdit(row.task)"
					@contextmenu.prevent="onTaskContextMenu(row.task)"
					@click="onRowClick(row, index)"
				>
					<template v-if="currentEditingId === row.task.id">
						<div class="task-inline-edit-row">
							<input
								ref="inlineEditInputRef"
								type="text"
								class="task-inline-input"
								:value="currentEditingDraft"
								@input="setEditingDraft($event.target.value)"
								@keydown.enter.prevent="onInlineEditEnter()"
								@keydown.escape.prevent="injectedCancelEdit()"
								@click.stop
							/>
							<button type="button" class="task-inline-save-btn" title="Save (Enter)" @click.stop="onInlineEditSave">Save</button>
							<button type="button" class="task-inline-cancel-btn" title="Cancel (Escape)" @click.stop="injectedCancelEdit()">Cancel</button>
						</div>
					</template>
					<template v-else>
						<span v-if="displayMode === 'hierarchy' && row.task.countSubitems > 0" class="task-project-indicator" :title="collapsedParentIds.includes(row.task.id) ? 'Expand subtasks' : 'Collapse subtasks'" @click.stop="toggleCollapse(row.task.id)">
							<HugeiconsIcon :icon="collapsedParentIds.includes(row.task.id) ? Folder01Icon : FolderOpenIcon" width="1.1em" height="1.1em" />
						</span>
						<span v-else-if="row.isNextAction" class="task-next-action-indicator" title="Next action for project">
							<HugeiconsIcon :icon="PlayIcon" width="1.1em" height="1.1em" />
						</span>
						<span v-else-if="isFutureWait(row.task)" class="task-wait-indicator" :title="'Waiting until ' + formatDateDisplay(row.task.waitUntil)">
							<HugeiconsIcon :icon="ClockIcon" width="1.1em" height="1.1em" />
						</span>
						<span class="task-content" :class="{ 'task-content-waiting': isFutureWait(row.task) }">{{ row.task.content }}</span>
						<span v-if="(row.task.priority || row.task.tags?.length || row.task.contexts?.length || row.task.dueDate)" class="task-meta">
							<span v-if="row.task.priority" class="task-priority" :title="'Priority ' + row.task.priority">{{ row.task.priority }}</span>
							<span v-for="tag in (row.task.tags || [])" :key="'tag-' + tag" class="tag task-meta-clickable" :style="tagStyle(tag)" title="Search for #{{ tag }}" @click.stop="setSearchQuery('#' + tag)">#{{ tag }}</span>
							<span v-for="ctx in (row.task.contexts || [])" :key="'ctx-' + ctx" class="context task-meta-clickable" :style="contextStyle(ctx)" title="Search for @{{ ctx }}" @click.stop="setSearchQuery('@' + ctx)">@{{ ctx }}</span>
							<span v-if="row.task.dueDate" class="task-due task-meta-clickable" :class="{ 'task-due-overdue': isOverdue(row.task) }" :title="'Due ' + formatDateDisplay(row.task.dueDate)">
								<HugeiconsIcon :icon="Calendar03Icon" width="1em" height="1em" />
								due {{ formatDateDisplay(row.task.dueDate) }}
							</span>
						</span>
					</template>
				</li>
			</ul>
		</div>
		<div v-if="listId && !searchQuery && (hiddenTagNames.length > 0 || hiddenContextNames.length > 0)" class="list-hidden-footer" role="status" aria-live="polite">
			<span class="list-hidden-label">Hidden by hide-at-times:</span>
			<span class="task-meta">
				<span v-for="tag in hiddenTagNames" :key="'ht-' + tag" class="tag task-meta-clickable" :style="tagStyle(tag)" :title="'Search for #' + tag" @click.stop="setSearchQuery('#' + tag)">#{{ tag }}</span>
				<span v-for="ctx in hiddenContextNames" :key="'hc-' + ctx" class="context task-meta-clickable" :style="contextStyle(ctx)" :title="'Search for @' + ctx" @click.stop="setSearchQuery('@' + ctx)">@{{ ctx }}</span>
			</span>
		</div>
		<div v-if="taskDetailsTask" class="confirm-overlay task-details-overlay" @click.self="closeTaskDetails">
			<div class="task-details-dialog" role="dialog" aria-labelledby="task-details-title" aria-modal="true" @click.stop>
				<h2 id="task-details-title" class="task-details-title">Task details</h2>
				<p class="task-details-content">{{ taskDetailsTask.content }}</p>
				<div class="task-details-due" v-if="taskDetailsTask?.dueDate">
					<label>Due date</label>
					<span class="task-details-due-value">{{ formatDateDisplay(taskDetailsTask.dueDate) }}</span>
					<button type="button" class="list-options-btn" @click="openDueDialogFromDetails">Change</button>
				</div>
				<div class="task-details-due" v-else>
					<label>Due date</label>
					<button type="button" class="list-options-btn" @click="openDueDialogFromDetails">Set due date</button>
				</div>
				<div class="task-details-priority">
					<label for="task-details-priority-slider">Priority: {{ prioritySliderLabel }}</label>
					<input
						id="task-details-priority-slider"
						v-model.number="prioritySliderValue"
						type="range"
						min="0"
						max="26"
						class="task-details-priority-slider"
					/>
				</div>
				<div class="task-details-notes">
					<label for="task-details-notes-input">Notes</label>
					<p v-if="taskNotesLoadError" class="task-details-notes-error" role="alert">Could not load notes.</p>
					<textarea id="task-details-notes-input" v-model="notesDraft" class="task-details-notes-textarea" rows="6" placeholder="Notes…"></textarea>
				</div>
				<div class="task-details-actions">
					<button type="button" class="task-details-done-btn good" title="Mark task done" aria-label="Mark task done" @click="openDoneConfirmFromDetails">
						<HugeiconsIcon :icon="CheckmarkBadge01Icon" width="1.25em" height="1.25em" />
					</button>
					<button type="button" class="task-details-close-btn" @click="closeTaskDetails">Close</button>
				</div>
			</div>
		</div>
		<div v-if="taskForWait" class="confirm-overlay task-details-overlay" @click.self="closeWaitDialog">
			<div class="task-details-dialog wait-dialog" role="dialog" aria-labelledby="wait-dialog-title" aria-modal="true" @click.stop>
				<h2 id="wait-dialog-title" class="task-details-title">Wait until</h2>
				<p class="task-details-content">{{ taskForWait.content }}</p>
				<div class="wait-dialog-field">
					<label for="wait-datetime-input">Date and time</label>
					<input id="wait-datetime-input" v-model="waitDateTimeDraft" type="datetime-local" class="wait-datetime-input" />
				</div>
				<div class="wait-dialog-shortcuts">
					<button type="button" class="wait-shortcut-btn" @click="setWaitRelativeDays(1)">+1 day</button>
					<button type="button" class="wait-shortcut-btn" @click="setWaitRelativeDays(2)">+2 days</button>
					<button type="button" class="wait-shortcut-btn" @click="setWaitRelativeDays(3)">+3 days</button>
				</div>
				<div class="task-details-actions wait-dialog-actions">
					<button type="button" class="task-details-close-btn" @click="closeWaitDialog">Cancel</button>
					<button type="button" class="task-details-close-btn" @click="clearWaitAndClose">Clear wait</button>
					<button type="button" class="task-details-close-btn good" @click="saveWaitAndClose">Save</button>
				</div>
			</div>
		</div>
		<div v-if="taskForDue" class="confirm-overlay task-details-overlay" @click.self="closeDueDialog">
			<div class="task-details-dialog wait-dialog" role="dialog" aria-labelledby="due-dialog-title" aria-modal="true" @click.stop>
				<h2 id="due-dialog-title" class="task-details-title">Due date</h2>
				<p class="task-details-content">{{ taskForDue.content }}</p>
				<div class="wait-dialog-field">
					<label for="due-date-input">Date</label>
					<input id="due-date-input" v-model="dueDateDraft" type="date" class="wait-datetime-input" />
				</div>
				<div class="wait-dialog-shortcuts">
					<button type="button" class="wait-shortcut-btn" @click="setDueRelativeDays(1)">+1 day</button>
					<button type="button" class="wait-shortcut-btn" @click="setDueRelativeDays(2)">+2 days</button>
					<button type="button" class="wait-shortcut-btn" @click="setDueRelativeDays(7)">+7 days</button>
				</div>
				<div class="task-details-actions wait-dialog-actions">
					<button type="button" class="task-details-close-btn" @click="closeDueDialog">Cancel</button>
					<button type="button" class="task-details-close-btn" @click="clearDueAndClose">Clear due</button>
					<button type="button" class="task-details-close-btn good" @click="saveDueAndClose">Save</button>
				</div>
			</div>
		</div>
		<div v-if="pendingDoneTask" class="confirm-overlay" @click.self="pendingDoneTask = null">
			<div class="confirm-dialog" role="dialog" aria-labelledby="confirm-title" aria-modal="true">
				<h2 id="confirm-title" class="confirm-title">Mark task done?</h2>
				<p class="confirm-message">{{ doneConfirmMessage }}</p>
				<div class="confirm-actions">
					<button type="button" class="confirm-btn confirm-cancel" @click="pendingDoneTask = null">Cancel</button>
					<button type="button" class="task-details-done-btn good" @click="confirmDone">
						Done
						<HugeiconsIcon :icon="CheckmarkBadge01Icon" width="1.25em" height="1.25em" />
					</button>
				</div>
			</div>
		</div>
		<div v-if="listOptionsOpen" class="list-options-overlay" @click.self="listOptionsOpen = false">
			<div class="list-options-dialog" role="dialog" aria-labelledby="list-options-title" aria-modal="true">
				<h2 id="list-options-title" class="list-options-title">List options</h2>
				<div class="list-options-rename">
					<label for="list-rename-input">Rename list</label>
					<input id="list-rename-input" v-model="listRenameTitle" type="text" class="list-rename-input" />
					<button type="button" class="list-options-action-btn" @click="renameList">Rename</button>
				</div>
				<div class="list-options-delete">
					<button type="button" class="list-options-delete-btn" @click="confirmDeleteList">Delete list</button>
					<p v-if="deleteListConfirm" class="list-options-delete-confirm">Tasks will be moved to Inbox. <button type="button" class="confirm-delete-list-btn" @click="deleteList">Confirm delete</button> <button type="button" class="cancel-delete-list-btn" @click="deleteListConfirm = false">Cancel</button></p>
				</div>
				<div class="list-options-dialog-actions">
					<button type="button" class="list-options-close" @click="closeListOptions">Close</button>
				</div>
			</div>
		</div>
	</Section>
</template>

<script setup>
	import { onMounted, onUnmounted, watch, computed, inject, nextTick } from 'vue';
	import { ref } from 'vue';
	import { useRouter } from 'vue-router';
	import { HugeiconsIcon } from '@hugeicons/vue';
	import { Folder01Icon, FolderOpenIcon, CheckmarkBadge01Icon, PlayIcon, ClockIcon, Calendar03Icon } from '@hugeicons/core-free-icons';
	import Section from 'picocrank/vue/components/Section.vue';
	import { useSettings } from '../composables/useSettings.js';

	const router = useRouter();
	const { formatDateDisplay } = useSettings();
	const items = ref([]);
	const allItems = ref([]);
	const hiddenTagNames = ref([]);
	const hiddenContextNames = ref([]);
	const loading = ref(false);
	const searchError = ref(null);
	const collapsedParentIds = ref([]);
	const pendingDoneTask = ref(null);
	const taskDetailsTask = ref(null);
	const notesDraft = ref('');
	const taskNotesLoadError = ref(false);
	const priorityDraft = ref('');
	const taskForWait = ref(null);
	// Slider 0 = none, 1–26 = A–Z
	const prioritySliderValue = ref(0);
	const prioritySliderLabel = computed(() => {
		const v = prioritySliderValue.value;
		if (v <= 0 || v > 26) return 'None';
		return String.fromCharCode(64 + v);
	});
	const waitDateTimeDraft = ref('');
	const taskForDue = ref(null);
	const dueDateDraft = ref('');
	const listOptionsOpen = ref(false);
	const listRenameTitle = ref('');
	const listTitle = ref('');
	const deleteListConfirm = ref(false);
	const inlineEditInputRef = ref(null);
	const injectedStartEdit = inject('startEdit', () => {});
	const editingTaskId = inject('editingTaskId', ref(null));
	const editingDraft = inject('editingDraft', ref(''));
	const injectedCancelEdit = inject('cancelEdit', () => {});
	const injectedSubmitEdit = inject('submitEdit', () => {});
	const refreshTrigger = inject('refreshTrigger', ref(0));
	const focusFirstListItemTrigger = inject('focusFirstListItemTrigger', ref(0));
	const setSearchQuery = inject('setSearchQuery', () => {});
	const displayMode = inject('displayMode', ref('hierarchy'));
	const focusAddTaskInput = inject('focusAddTaskInput', () => {});
	const showToast = inject('showToast', () => {});
	const taskPropertyProperties = inject('taskPropertyProperties', ref({ tagProperties: {}, contextProperties: {} }));

	const focusedIndex = ref(0);
	const ddLastKey = ref(null);
	const ddLastTime = ref(0);
	const DD_TIMEOUT_MS = 500;
	const yankedTask = ref(null);

	const sectionTitle = computed(() => {
		if (props.searchQuery) return 'Search results';
		return listTitle.value || 'Tasks';
	});
	const currentEditingId = computed(() => editingTaskId?.value ?? null);
	const currentEditingDraft = computed(() => editingDraft?.value ?? '');
	function setEditingDraft(v) {
		if (editingDraft) editingDraft.value = v;
	}
	function onInlineEditSave() {
		const content = (currentEditingDraft.value ?? '').trim();
		if (!content) {
			showToast('Task content cannot be empty', 'error');
			return;
		}
		injectedSubmitEdit();
	}
	function onInlineEditEnter() {
		onInlineEditSave();
	}

	const props = defineProps({
		listId: {
			type: String,
			default: null,
		},
		searchQuery: {
			type: String,
			default: '',
		},
	});

	function defaultPastelStyle(name) {
		let h = 0;
		const s = String(name || '');
		for (let i = 0; i < s.length; i++) {
			h = ((h << 5) - h) + s.charCodeAt(i);
			h |= 0;
		}
		const hue = ((h % 360) + 360) % 360;
		return {
			backgroundColor: `hsl(${hue}, 65%, 88%)`,
			color: `hsl(${hue}, 45%, 35%)`,
		};
	}

	function tagStyle(name) {
		const props = taskPropertyProperties?.value?.tagProperties?.[name]?.props;
		if (!props) return defaultPastelStyle(name);
		const css = props.css?.trim();
		if (css) return css;
		const bg = props.bgcolor;
		if (bg) return { backgroundColor: bg, color: '#333' };
		return defaultPastelStyle(name);
	}

	function contextStyle(name) {
		const props = taskPropertyProperties?.value?.contextProperties?.[name]?.props;
		if (!props) return defaultPastelStyle(name);
		const css = props.css?.trim();
		if (css) return css;
		const bg = props.bgcolor;
		if (bg) return { backgroundColor: bg, color: '#333' };
		return defaultPastelStyle(name);
	}

	function isFutureWait(task) {
		if (!task?.waitUntil) return false;
		const t = new Date(task.waitUntil);
		return !isNaN(t.getTime()) && t > new Date();
	}

	function getFilteredByCollapsed() {
		const full = allItems.value;
		const collapsed = collapsedParentIds.value;
		if (full.length === 0) return [];
		const idToTask = new Map();
		for (const row of full) {
			idToTask.set(row.task.id, row.task);
		}
		return full.filter((row) => {
			let id = row.task.parentId;
			while (id) {
				if (collapsed.includes(id)) return false;
				const parent = idToTask.get(id);
				id = parent?.parentId ?? '';
			}
			return true;
		});
	}

	function getOnlyNextActionList() {
		const full = allItems.value.filter((row) => !isFutureWait(row.task));
		if (full.length === 0) return [];
		const idToRow = new Map();
		for (const row of full) {
			idToRow.set(row.task.id, row);
		}
		function findFirstAction(taskId) {
			const row = idToRow.get(taskId);
			if (!row) return null;
			if (row.task.countSubitems === 0) return row.task;
			const children = full.filter((r) => r.task.parentId === taskId);
			if (children.length === 0) return null;
			const firstChild = children[0];
			return findFirstAction(firstChild.task.id);
		}
		const result = [];
		const roots = full.filter((r) => r.depth === 0);
		for (const rootRow of roots) {
			if (rootRow.task.countSubitems === 0) {
				result.push({ task: rootRow.task, depth: 0, isNextAction: false });
			} else {
				const firstAction = findFirstAction(rootRow.task.id);
				if (firstAction) {
					result.push({ task: firstAction, depth: 0, isNextAction: true });
				}
			}
		}
		return result;
	}

	function getOnlyWaitingList() {
		return allItems.value
			.filter((row) => isFutureWait(row.task))
			.map((row) => ({ ...row, depth: 0 }));
	}

	function updateItems() {
		if (props.searchQuery) return;
		if (!allItems.value.length) {
			items.value = [];
			return;
		}
		if (displayMode.value === 'onlyNextAction') {
			items.value = getOnlyNextActionList();
		} else if (displayMode.value === 'onlyWaiting') {
			items.value = getOnlyWaitingList();
		} else {
			items.value = getFilteredByCollapsed();
		}
	}

	async function loadListTasks() {
		if (!props.listId || !window.client) return;
		const ret = await window.client.listTasks({
			parentType: 'list',
			parentId: props.listId,
		});
		const tasks = ret.tasks || [];
		const tree = ret.tree || {};
		hiddenTagNames.value = ret.hiddenTagNames || [];
		hiddenContextNames.value = ret.hiddenContextNames || [];
		const idToTask = new Map(tasks.map((t) => [t.id, t]));
		const flat = [];
		function walk(parentId, depth) {
			const childIds = tree[parentId]?.ids ?? [];
			for (const childId of childIds) {
				const task = idToTask.get(childId);
				if (task) {
					flat.push({ task, depth });
					walk(childId, depth + 1);
				}
			}
		}
		walk(props.listId, 0);
		allItems.value = flat;
		updateItems();
	}

	function toggleCollapse(taskId) {
		const c = collapsedParentIds.value;
		if (c.includes(taskId)) {
			collapsedParentIds.value = c.filter((x) => x !== taskId);
		} else {
			collapsedParentIds.value = [...c, taskId];
		}
		updateItems();
	}

	function onRowClick(row, index) {
		if (row.task.countSubitems > 0) {
			toggleCollapse(row.task.id);
		}
		focusedIndex.value = index;
	}

	async function loadSearchResults() {
		const q = (props.searchQuery || '').trim();
		searchError.value = null;
		if (!q || !window.client) {
			items.value = [];
			return;
		}
		try {
			const res = await window.client.searchTasks({ query: q });
			items.value = (res.tasks || []).map((t) => ({ task: t, depth: 0 }));
		} catch (e) {
			const reason = e?.message || String(e);
			searchError.value = reason;
			items.value = [];
			showToast('Search failed. Please try again.', 'error');
		}
	}

	async function load() {
		loading.value = true;
		if (!props.searchQuery?.trim()) {
			searchError.value = null;
		}
		try {
			if (props.searchQuery?.trim()) {
				await loadSearchResults();
			} else if (props.listId) {
				await loadListTasks();
			} else {
				items.value = [];
				hiddenTagNames.value = [];
				hiddenContextNames.value = [];
			}
		} finally {
			loading.value = false;
		}
	}

	function onListKeydown(e) {
		if (taskDetailsTask.value && e.key === 'Escape') {
			e.preventDefault();
			closeTaskDetails();
			return;
		}
		if (taskForWait.value && e.key === 'Escape') {
			e.preventDefault();
			closeWaitDialog();
			return;
		}
		if (taskForDue.value && e.key === 'Escape') {
			e.preventDefault();
			closeDueDialog();
			return;
		}
		if (e.target.closest('input, textarea, select, [contenteditable="true"]')) return;
		if (listOptionsOpen.value || pendingDoneTask.value || taskDetailsTask.value || taskForWait.value || taskForDue.value) return;
		if (currentEditingId.value) return;

		const len = items.value.length;

		if (e.key === 'Escape') {
			e.preventDefault();
			focusedIndex.value = -1;
			return;
		}

		if (len === 0) {
			if (e.key === 'i' && !e.ctrlKey && !e.metaKey && !e.altKey) {
				e.preventDefault();
				focusAddTaskInput(null);
			}
			return;
		}

		if (e.key === 'j' || e.key === 'ArrowDown') {
			e.preventDefault();
			ddLastKey.value = null;
			focusedIndex.value = focusedIndex.value < 0 ? 0 : Math.min(focusedIndex.value + 1, len - 1);
			return;
		}
		if (e.key === 'k' || e.key === 'ArrowUp') {
			e.preventDefault();
			ddLastKey.value = null;
			focusedIndex.value = focusedIndex.value < 0 ? -1 : Math.max(0, focusedIndex.value - 1);
			return;
		}
		if (e.key === 'Enter' || e.key === 'e' || e.key === 'F2') {
			e.preventDefault();
			ddLastKey.value = null;
			const row = items.value[focusedIndex.value];
			if (row) injectedStartEdit(row.task);
			return;
		}
		if (e.key === 'd') {
			if (ddLastKey.value === 'd' && Date.now() - ddLastTime.value < DD_TIMEOUT_MS) {
				e.preventDefault();
				ddLastKey.value = null;
				const row = items.value[focusedIndex.value];
				if (row) {
					pendingDoneTask.value = row.task;
				}
				return;
			}
			ddLastKey.value = 'd';
			ddLastTime.value = Date.now();
			return;
		}
		if (e.key === 'Delete' && focusedIndex.value >= 0) {
			e.preventDefault();
			ddLastKey.value = null;
			const row = items.value[focusedIndex.value];
			if (row) pendingDoneTask.value = row.task;
			return;
		}
		if (e.key === 'y') {
			if (ddLastKey.value === 'y' && Date.now() - ddLastTime.value < DD_TIMEOUT_MS) {
				e.preventDefault();
				ddLastKey.value = null;
				const row = items.value[focusedIndex.value];
				if (row) yankedTask.value = { id: row.task.id, content: row.task.content };
				return;
			}
			ddLastKey.value = 'y';
			ddLastTime.value = Date.now();
			return;
		}
		if ((e.key === 'w' || e.key === 'W') && focusedIndex.value >= 0) {
			e.preventDefault();
			ddLastKey.value = null;
			const row = items.value[focusedIndex.value];
			if (row) openWaitDialog(row.task);
			return;
		}
		if ((e.key === 'u' || e.key === 'U') && focusedIndex.value >= 0) {
			e.preventDefault();
			ddLastKey.value = null;
			const row = items.value[focusedIndex.value];
			if (row) openDueDialog(row.task);
			return;
		}
		if (e.key === 'p') {
			e.preventDefault();
			ddLastKey.value = null;
			if (yankedTask.value && window.client) {
				const row = items.value[focusedIndex.value];
				const parentTaskId = row ? row.task.id : '';
				window.client.createTask({
					content: yankedTask.value.content,
					parentListId: props.listId || '',
					parentTaskId,
				}).then(() => {
					if (refreshTrigger) refreshTrigger.value++;
					showToast('Task added');
				}).catch((err) => {
					const reason = err?.message || String(err);
					showToast('Could not add task: ' + reason, 'error');
					// yanked content kept so user can retry
				});
			}
			return;
		}
		if (e.key === 'r' && len > 0) {
			e.preventDefault();
			ddLastKey.value = null;
			const prevRow = items.value[focusedIndex.value];
			const arr = [...items.value];
			for (let i = arr.length - 1; i > 0; i--) {
				const j = Math.floor(Math.random() * (i + 1));
				[arr[i], arr[j]] = [arr[j], arr[i]];
			}
			items.value = arr;
			if (prevRow) {
				const newIndex = items.value.findIndex((r) => r.task.id === prevRow.task.id);
				if (newIndex >= 0) focusedIndex.value = newIndex;
			}
			return;
		}
		ddLastKey.value = null;
		if (e.key === 'i' && !e.ctrlKey && !e.metaKey && !e.altKey) {
			e.preventDefault();
			const row = items.value[focusedIndex.value];
			focusAddTaskInput(row ? row.task.id : null);
		}
	}

	async function fetchListTitle() {
		if (!props.listId || !window.client) return;
		const res = await window.client.getLists({});
		const list = (res.lists || []).find((l) => l.id === props.listId);
		listTitle.value = list?.title ?? '';
		listRenameTitle.value = listTitle.value;
	}

	function openListOptions() {
		listRenameTitle.value = listTitle.value;
		deleteListConfirm.value = false;
		listOptionsOpen.value = true;
	}

	function closeListOptions() {
		listOptionsOpen.value = false;
	}

	async function renameList() {
		const title = (listRenameTitle.value || '').trim();
		if (!title || !props.listId || !window.client) return;
		try {
			await window.client.updateList({ id: props.listId, title });
			listTitle.value = title;
			closeListOptions();
			if (refreshTrigger) refreshTrigger.value++;
			showToast('List renamed');
		} catch (e) {
			const reason = e?.message || String(e);
			showToast('Could not rename list: ' + reason, 'error');
			// keep dialog open so user can retry
		}
	}

	function confirmDeleteList() {
		deleteListConfirm.value = true;
	}

	async function deleteList() {
		if (!props.listId || !window.client) return;
		try {
			await window.client.deleteList({ id: props.listId });
			closeListOptions();
			router.push('/');
			if (refreshTrigger) refreshTrigger.value++;
			showToast('List deleted');
		} catch (e) {
			const reason = e?.message || String(e);
			showToast('Could not delete list: ' + reason, 'error');
			// keep dialog open so user can retry
		}
		deleteListConfirm.value = false;
	}

	async function doneTask(id) {
		if (!id || !window.client) return;
		try {
			await window.client.doneTask({ id });
			if (currentEditingId.value === id) injectedCancelEdit();
			load();
			showToast('Task completed');
			pendingDoneTask.value = null;
		} catch (e) {
			const reason = e?.message || String(e);
			showToast('Could not complete task: ' + reason, 'error');
			// keep confirm dialog open so user can retry
		}
	}

	const doneConfirmMessage = computed(() => {
		if (!pendingDoneTask.value) return '';
		const c = pendingDoneTask.value.content || '';
		return c ? `"${c.length > 60 ? c.slice(0, 60) + '…' : c}"` : 'This task';
	});

	function onTaskContextMenu(item) {
		taskDetailsTask.value = item;
	}

	function priorityLetterToSlider(letter) {
		const p = (letter ?? '').trim().toUpperCase();
		if (!p || p.length === 0) return 0;
		const code = p.charCodeAt(0);
		if (code >= 65 && code <= 90) return code - 64; // A=1, Z=26
		return 0;
	}

	async function loadTaskNotes() {
		if (!taskDetailsTask.value?.id || !window.client) return;
		priorityDraft.value = (taskDetailsTask.value.priority ?? '').trim().toUpperCase().slice(0, 1) || '';
		taskNotesLoadError.value = null;
		try {
			const res = await window.client.getTaskMetadata({ taskId: taskDetailsTask.value.id });
			notesDraft.value = res.fields?.notes ?? '';
			if (res.fields?.priority != null) {
				const p = (res.fields.priority ?? '').trim().toUpperCase().slice(0, 1);
				priorityDraft.value = p;
			}
		} catch (e) {
			const reason = e?.message || String(e);
			showToast('Could not load task details: ' + reason, 'error');
			taskNotesLoadError.value = true;
			// do not overwrite notesDraft so user does not save over real notes
		}
		prioritySliderValue.value = priorityLetterToSlider(priorityDraft.value);
	}

	async function closeTaskDetails() {
		if (taskDetailsTask.value?.id && window.client) {
			try {
				await saveTaskNotes();
				await saveTaskPriority();
			} catch (e) {
				showToast('Could not save: ' + (e?.message || String(e)), 'error');
				return;
			}
		}
		taskDetailsTask.value = null;
		notesDraft.value = '';
		taskNotesLoadError.value = false;
		priorityDraft.value = '';
		prioritySliderValue.value = 0;
	}

	async function saveTaskNotes() {
		if (!taskDetailsTask.value?.id || !window.client) return;
		await window.client.setTaskMetadata({
			taskId: taskDetailsTask.value.id,
			field: 'notes',
			value: notesDraft.value ?? '',
		});
	}

	async function saveTaskPriority() {
		if (!taskDetailsTask.value?.id || !window.client) return;
		const v = prioritySliderValue.value;
		const val = v >= 1 && v <= 26 ? String.fromCharCode(64 + v) : '';
		await window.client.setTaskMetadata({
			taskId: taskDetailsTask.value.id,
			field: 'priority',
			value: val,
		});
		if (refreshTrigger) refreshTrigger.value++;
	}

	function formatWaitForInput(isoStr) {
		if (!isoStr) return '';
		const d = new Date(isoStr);
		if (isNaN(d.getTime())) return '';
		const y = d.getFullYear();
		const m = String(d.getMonth() + 1).padStart(2, '0');
		const day = String(d.getDate()).padStart(2, '0');
		const h = String(d.getHours()).padStart(2, '0');
		const min = String(d.getMinutes()).padStart(2, '0');
		return `${y}-${m}-${day}T${h}:${min}`;
	}

	function openWaitDialog(task) {
		taskForWait.value = task;
		waitDateTimeDraft.value = formatWaitForInput(task.waitUntil);
	}

	function setWaitRelativeDays(days) {
		const d = new Date();
		d.setDate(d.getDate() + days);
		waitDateTimeDraft.value = formatWaitForInput(d.toISOString());
	}

	function closeWaitDialog() {
		taskForWait.value = null;
		waitDateTimeDraft.value = '';
	}

	async function saveWaitAndClose() {
		if (!taskForWait.value?.id || !window.client) {
			closeWaitDialog();
			return;
		}
		const val = waitDateTimeDraft.value?.trim();
		const iso = val ? new Date(val).toISOString() : '';
		try {
			await window.client.setTaskMetadata({
				taskId: taskForWait.value.id,
				field: 'wait',
				value: iso,
			});
			if (refreshTrigger) refreshTrigger.value++;
			showToast('Wait until saved');
			closeWaitDialog();
		} catch (e) {
			showToast('Could not save wait until: ' + (e?.message || String(e)), 'error');
			// keep dialog open so user can retry
		}
	}

	async function clearWaitAndClose() {
		if (!taskForWait.value?.id || !window.client) {
			closeWaitDialog();
			return;
		}
		try {
			await window.client.setTaskMetadata({
				taskId: taskForWait.value.id,
				field: 'wait',
				value: '',
			});
			if (refreshTrigger) refreshTrigger.value++;
			closeWaitDialog();
		} catch (e) {
			showToast('Could not clear wait until: ' + (e?.message || String(e)), 'error');
			// keep dialog open so user can retry
		}
	}

	function formatDueForInput(isoStr) {
		if (!isoStr) return '';
		const d = new Date(isoStr);
		if (isNaN(d.getTime())) return '';
		const y = d.getFullYear();
		const m = String(d.getMonth() + 1).padStart(2, '0');
		const day = String(d.getDate()).padStart(2, '0');
		return `${y}-${m}-${day}`;
	}

	function isOverdue(task) {
		if (!task?.dueDate) return false;
		const d = new Date((task.dueDate || '').includes('T') ? task.dueDate : task.dueDate + 'T23:59:59');
		if (isNaN(d.getTime())) return false;
		return d < new Date();
	}

	function openDueDialog(task) {
		taskForDue.value = task;
		dueDateDraft.value = formatDueForInput(task.dueDate);
	}

	function openDueDialogFromDetails() {
		if (taskDetailsTask.value) {
			taskForDue.value = taskDetailsTask.value;
			dueDateDraft.value = formatDueForInput(taskDetailsTask.value.dueDate ?? '');
		}
	}

	function setDueRelativeDays(days) {
		const d = new Date();
		d.setDate(d.getDate() + days);
		dueDateDraft.value = formatDueForInput(d.toISOString());
	}

	function closeDueDialog() {
		taskForDue.value = null;
		dueDateDraft.value = '';
	}

	async function saveDueAndClose() {
		if (!taskForDue.value?.id || !window.client) {
			closeDueDialog();
			return;
		}
		const val = (dueDateDraft.value ?? '').trim();
		try {
			await window.client.setTaskMetadata({
				taskId: taskForDue.value.id,
				field: 'due',
				value: val,
			});
			if (refreshTrigger) refreshTrigger.value++;
			showToast('Due date saved');
			closeDueDialog();
		} catch (e) {
			showToast('Could not save due date: ' + (e?.message || String(e)), 'error');
			// keep dialog open so user can retry
		}
	}

	async function clearDueAndClose() {
		if (!taskForDue.value?.id || !window.client) {
			closeDueDialog();
			return;
		}
		try {
			await window.client.setTaskMetadata({
				taskId: taskForDue.value.id,
				field: 'due',
				value: '',
			});
			if (refreshTrigger) refreshTrigger.value++;
			closeDueDialog();
		} catch (e) {
			showToast('Could not clear due date: ' + (e?.message || String(e)), 'error');
			// keep dialog open so user can retry
		}
	}

	function openDoneConfirmFromDetails() {
		if (taskDetailsTask.value) {
			pendingDoneTask.value = taskDetailsTask.value;
			closeTaskDetails();
		}
	}

	function confirmDone() {
		if (pendingDoneTask.value) {
			doneTask(pendingDoneTask.value.id);
			// pendingDoneTask cleared in doneTask on success; on error we keep dialog open
		}
	}

	onMounted(() => {
		load();
		document.addEventListener('keydown', onListKeydown);
	});
	onUnmounted(() => {
		document.removeEventListener('keydown', onListKeydown);
	});
	watch([() => props.listId, () => props.searchQuery], load);
	watch(() => props.listId, (id) => {
		if (id) fetchListTitle();
	}, { immediate: true });
	watch(refreshTrigger, load);
	watch(displayMode, () => { if (!props.searchQuery) updateItems(); });
	watch([allItems, collapsedParentIds], () => { if (!props.searchQuery && allItems.value.length) updateItems(); });
	watch(taskDetailsTask, (task) => {
		if (task) loadTaskNotes();
		else {
			notesDraft.value = '';
			taskNotesLoadError.value = false;
			priorityDraft.value = '';
			prioritySliderValue.value = 0;
		}
	});
	watch(focusFirstListItemTrigger, () => {
		if (items.value.length > 0) focusedIndex.value = 0;
	});
	watch(items, (newItems) => {
		focusedIndex.value = Math.min(focusedIndex.value, Math.max(0, newItems.length - 1));
	}, { immediate: true });
	watch(currentEditingId, (id) => {
		if (id) nextTick(() => (Array.isArray(inlineEditInputRef.value) ? inlineEditInputRef.value[0] : inlineEditInputRef.value)?.focus());
	});
</script>

<style scoped>
	.list-loading,
	.list-empty {
		text-align: center;
		padding: 1.5rem 1rem;
	}
	.list-loading {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
	}
	.list-loading p,
	.list-empty p {
		margin: 0;
		color: #555;
		font-size: 0.95rem;
	}
	.list-search-error {
		padding: 1.25rem 1rem;
		text-align: center;
		background: #fff5f5;
		border: 1px solid #feb2b2;
		border-radius: 0.35rem;
		margin: 0 0.25rem;
	}
	.list-search-error-title {
		margin: 0 0 0.35rem 0;
		font-weight: 600;
		color: #c00;
		font-size: 0.95rem;
	}
	.list-search-error-reason {
		margin: 0;
		color: #666;
		font-size: 0.875rem;
		word-break: break-word;
	}
	.list-loading-spinner {
		display: inline-block;
		width: 1.5rem;
		height: 1.5rem;
		border: 2px solid #e0e0e0;
		border-top-color: var(--femtocrank-primary, #2dae82);
		border-radius: 50%;
		animation: list-loading-spin 0.7s linear infinite;
	}
	@keyframes list-loading-spin {
		to { transform: rotate(360deg); }
	}
	.search-heading {
		margin-bottom: 1rem;
		padding: 0 0.25rem;
	}
	.search-heading p {
		margin: 0;
		font-size: 0.95rem;
		color: #555;
	}

	li.task-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		background-color: white;
		padding: 0.5rem;
		border-radius: 0;
		margin-bottom: 0.2em;
		box-shadow: none;
		border-bottom: 1px solid #e5e5e5;
		cursor: pointer;
	}
	li.task-row:hover {
		background-color: #f9f9f9;
	}
	li.task-row-project {
		background-color: #fafafa;
	}
	li.task-row-project:hover {
		background-color: #f0f0f0;
	}
	li.task-row-editing {
		background-color: #f0f6ff;
		box-shadow: inset 0 0 0 1px rgba(59, 130, 246, 0.25);
	}
	li.task-row-focused {
		box-shadow: 0 0 0 2px #000;
		outline: none;
	}
	li.task-row-yanked {
		border-left: 3px solid #2563eb;
		background-color: #eff6ff;
	}

	.task-project-indicator {
		display: inline-flex;
		flex-shrink: 0;
		color: #c00;
		margin-right: 0.35rem;
		cursor: pointer;
	}
	.task-next-action-indicator {
		display: inline-flex;
		flex-shrink: 0;
		color: #0a7c42;
		margin-right: 0.35rem;
	}
	.task-wait-indicator {
		display: inline-flex;
		flex-shrink: 0;
		color: #6b7280;
		margin-right: 0.35rem;
	}
	.task-content-waiting {
		color: #9ca3af;
	}
	li.task-row-waiting .task-content {
		color: #9ca3af;
	}
	.task-content {
		flex: 1;
		min-width: 0;
	}

	.task-inline-edit-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex: 1;
		min-width: 0;
	}
	.task-inline-input {
		flex: 1;
		min-width: 0;
		padding: 0.35rem 0.5rem;
		border: 1px solid rgba(59, 130, 246, 0.5);
		border-radius: 0.35rem;
		font-size: inherit;
		background: #fff;
	}
	.task-inline-input:focus {
		outline: none;
		border-color: rgba(59, 130, 246, 0.8);
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15);
	}
	.task-inline-save-btn,
	.task-inline-cancel-btn {
		flex-shrink: 0;
		padding: 0.35rem 0.6rem;
		border-radius: 0.35rem;
		font-size: 0.875rem;
		cursor: pointer;
		border: 1px solid #ccc;
		background: #fff;
	}
	.task-inline-save-btn {
		border-color: var(--femtocrank-primary, #2dae82);
		background: var(--femtocrank-primary, #2dae82);
		color: #fff;
	}
	.task-inline-save-btn:hover {
		filter: brightness(1.05);
	}
	.task-inline-cancel-btn:hover {
		background: #f0f0f0;
	}

	.task-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem;
		justify-content: flex-end;
		flex-shrink: 0;
	}

	.task-meta span {
		padding: 0.3rem 0.6rem;
	}
	.task-meta-clickable {
		cursor: pointer;
	}

	.task-meta .task-priority {
		font-size: 0.85em;
		font-weight: 600;
		color: #555;
		background: #e8e8e8;
		border-radius: 0.25rem;
		padding: 0.2rem 0.45rem;
		margin-bottom: 0;
	}

	.task-meta .tag {
		font-size: 0.85em;
		border-radius: 0.25rem;
		margin-bottom: 0;
	}

	.task-meta .context {
		font-weight: bold;
		color: #2c5282;
		background: #c6d0d7;
		border-radius: 0.25rem;
	}
	.task-meta .task-due {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		font-size: 0.85em;
		color: #555;
		background: #e8f4e8;
		border-radius: 0.25rem;
		padding: 0.2rem 0.45rem;
	}
	.task-meta .task-due-overdue {
		background: #ffe8e8;
		color: #a00;
	}
	.task-details-due {
		margin-bottom: 0.5rem;
	}
	.task-details-due label {
		display: block;
		margin-bottom: 0.35rem;
		font-size: 0.9rem;
		color: #555;
	}
	.task-details-due-value {
		margin-right: 0.5rem;
	}
	.task-details-due .list-options-btn {
		margin-top: 0.25rem;
	}

	ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.confirm-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.4);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}
	.task-details-dialog {
		background: #fff;
		border-radius: 0.5rem;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
		padding: 1.25rem;
		min-width: 22rem;
		max-width: 90vw;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.task-details-title {
		margin: 0;
		font-size: 1.1rem;
	}
	.task-details-content {
		margin: 0;
		color: #333;
		font-size: 0.95rem;
		line-height: 1.4;
		padding: 0.5rem 0;
		border-bottom: 1px solid #eee;
	}
	.task-details-priority {
		margin-bottom: 0.5rem;
	}
	.task-details-priority label {
		display: block;
		margin-bottom: 0.35rem;
		font-size: 0.9rem;
		color: #555;
	}
	.task-details-priority-slider {
		width: 100%;
		max-width: 14rem;
		height: 0.5rem;
		margin-top: 0.25rem;
		accent-color: var(--femtocrank-primary, #2dae82);
		cursor: pointer;
	}
	.task-details-notes label {
		display: block;
		margin-bottom: 0.35rem;
		font-size: 0.9rem;
		color: #555;
	}
	.task-details-notes-error {
		margin: 0 0 0.35rem 0;
		font-size: 0.875rem;
		color: #c00;
	}
	.task-details-notes-textarea {
		width: 100%;
		box-sizing: border-box;
		padding: 0.5rem;
		border: 1px solid #ccc;
		border-radius: 0.35rem;
		font-size: 0.95rem;
		font-family: inherit;
		resize: vertical;
		min-height: 6em;
	}
	.task-details-actions {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-top: 0.25rem;
	}
	.task-details-done-btn {
		padding: 0.35rem 0.5rem;
		border: none;
		border-radius: 0.35rem;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.35rem;
	}
	.task-details-close-btn {
		padding: 0.4rem 0.75rem;
		border-radius: 0.35rem;
		border: 1px solid #ccc;
		font-size: 0.9rem;
		cursor: pointer;
	}
	.task-details-close-btn:hover {
		background: #f5f5f5;
	}
	.wait-dialog-field label {
		display: block;
		margin-bottom: 0.35rem;
		font-size: 0.9rem;
		color: #555;
	}
	.wait-datetime-input {
		width: 100%;
		box-sizing: border-box;
		padding: 0.4rem 0.5rem;
		border: 1px solid #ccc;
		border-radius: 0.35rem;
		font-size: 0.95rem;
	}
	.wait-dialog-actions .task-details-close-btn:last-child {
		margin-left: auto;
	}
	.wait-dialog-shortcuts {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
	}
	.wait-shortcut-btn {
		padding: 0.35rem 0.6rem;
		border: 1px solid #ccc;
		border-radius: 0.35rem;
		background: #fff;
		font-size: 0.9rem;
		cursor: pointer;
	}
	.wait-shortcut-btn:hover {
		background: #f5f5f5;
	}
	.confirm-dialog {
		background: #fff;
		border-radius: 0.5rem;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
		padding: 1.25rem;
		min-width: 18rem;
		max-width: 90vw;
	}
	.confirm-title {
		margin: 0 0 0.5rem 0;
		font-size: 1.1rem;
	}
	.confirm-message {
		margin: 0 0 1.25rem 0;
		color: #555;
		font-size: 0.95rem;
	}
	.confirm-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
	}
	.confirm-btn {
		padding: 0.4rem 0.75rem;
		border-radius: 0.35rem;
		border: 1px solid #ccc;
		font-size: 0.9rem;
		cursor: pointer;
		background: #fff;
	}
	.confirm-btn.confirm-cancel:hover {
		background: #f0f0f0;
	}

	.list-hidden-footer {
		margin-top: 1rem;
		padding: 0.5rem 0.25rem;
		font-size: 0.875rem;
		color: var(--femtocrank-text-muted, #666);
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.35rem;
	}
	.list-hidden-label {
		margin-right: 0.25rem;
	}
	.list-options-btn {
		padding: 0.4rem 0.75rem;
		border: 1px solid #ccc;
		border-radius: 0.35rem;
		background: #fff;
		font-size: 0.9rem;
		cursor: pointer;
	}
	.list-options-btn:hover {
		background: #f5f5f5;
	}

	.list-options-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.4);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}
	.list-options-dialog {
		background: #fff;
		border-radius: 0.5rem;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
		padding: 1.25rem;
		min-width: 20rem;
		max-width: 90vw;
	}
	.list-options-title {
		margin: 0 0 1rem 0;
		font-size: 1.1rem;
	}
	.list-options-rename {
		margin-bottom: 1rem;
	}
	.list-options-rename label {
		display: block;
		margin-bottom: 0.35rem;
		font-size: 0.9rem;
		color: #555;
	}
	.list-rename-input {
		width: 100%;
		box-sizing: border-box;
		padding: 0.4rem 0.5rem;
		border: 1px solid #ccc;
		border-radius: 0.35rem;
		font-size: 0.95rem;
		margin-bottom: 0.5rem;
	}
	.list-options-action-btn {
		padding: 0.35rem 0.6rem;
		border: 1px solid #ccc;
		border-radius: 0.35rem;
		background: #fff;
		font-size: 0.9rem;
		cursor: pointer;
	}
	.list-options-action-btn:hover {
		background: #f0f0f0;
	}
	.list-options-delete {
		margin-bottom: 1rem;
	}
	.list-options-delete-btn {
		padding: 0.35rem 0.6rem;
		border: 1px solid #c00;
		border-radius: 0.35rem;
		background: #fff;
		color: #c00;
		font-size: 0.9rem;
		cursor: pointer;
	}
	.list-options-delete-btn:hover {
		background: #ffe0e0;
	}
	.list-options-delete-confirm {
		margin: 0.5rem 0 0 0;
		font-size: 0.85rem;
		color: #555;
	}
	.confirm-delete-list-btn {
		margin-left: 0.25rem;
		padding: 0.2rem 0.4rem;
		background: #c00;
		color: #fff;
		border: none;
		border-radius: 0.25rem;
		cursor: pointer;
		font-size: 0.85rem;
	}
	.cancel-delete-list-btn {
		margin-left: 0.25rem;
		padding: 0.2rem 0.4rem;
		background: #eee;
		border: 1px solid #ccc;
		border-radius: 0.25rem;
		cursor: pointer;
		font-size: 0.85rem;
	}
	.list-options-dialog-actions {
		display: flex;
		justify-content: flex-end;
	}
	.list-options-close {
		padding: 0.4rem 0.75rem;
		border: 1px solid #ccc;
		border-radius: 0.35rem;
		background: #fff;
		font-size: 0.9rem;
		cursor: pointer;
	}
	.list-options-close:hover {
		background: #f0f0f0;
	}
</style>
