<template>
	<section class="welcome">
		<h2>Welcome</h2>
		<p class="welcome-intro">Choose a list to view or add tasks.</p>
		<div v-if="lists.length" class="list-nav">
			<router-link
				v-for="list in lists"
				:key="list.id"
				:to="'/lists/' + list.id"
				class="list-nav-btn"
			>
				{{ list.title }}
				<span v-if="list.countItems != null" class="list-count">{{ list.countItems }}</span>
			</router-link>
		</div>
		<p v-else-if="loadError" class="welcome-error" role="alert">
			{{ loadError }}
			<button type="button" class="welcome-retry-btn" @click="loadLists">Try again</button>
		</p>
		<p v-else-if="!loading" class="welcome-empty">No lists yet.</p>
		<p v-else class="welcome-loading">Loading…</p>
		<div class="welcome-actions">
			<button type="button" class="new-list-btn" :disabled="!isOnline" @click="showCreateDialog = true">
				New list
			</button>
		</div>
		<CreateListDialog
			v-model="showCreateDialog"
			@created="onListCreated"
		/>
	</section>
</template>

<script setup>
	import { ref, onMounted, watch, inject } from 'vue';
	import CreateListDialog from '../components/CreateListDialog.vue';
	import { getCachedInbox, INBOX_LIST_ID } from '../../../js/modules/offlineStorage.js';

	const lists = ref([]);
	const loading = ref(true);
	const loadError = ref(null);
	const showCreateDialog = ref(false);
	const refreshTrigger = inject('refreshTrigger', null);
	const showToast = inject('showToast', () => {});
	const isOnline = inject('isOnline', ref(true));

	async function loadLists() {
		loading.value = true;
		loadError.value = null;
		lists.value = [];
		if (!isOnline.value) {
			const cached = getCachedInbox();
			lists.value = cached
				? [{ id: cached.listId, title: cached.listTitle, countItems: cached.tasks?.length ?? null }]
				: [{ id: INBOX_LIST_ID, title: 'Inbox', countItems: null }];
			loading.value = false;
			return;
		}
		if (!window.client) {
			loading.value = false;
			return;
		}
		try {
			const res = await window.client.getLists();
			lists.value = res.lists || [];
		} catch (e) {
			const reason = e?.message || String(e);
			loadError.value = 'Could not load lists. Please try again.';
			showToast('Could not load lists: ' + reason, 'error');
		} finally {
			loading.value = false;
		}
	}

	function onListCreated() {
		loadLists();
		if (refreshTrigger) refreshTrigger.value++;
	}

	watch(isOnline, loadLists);
	onMounted(loadLists);
</script>

<style scoped>
	.welcome {
		padding: 1.5rem;
	}
	.welcome h2 {
		margin: 0 0 0.5rem 0;
		font-size: 1.5rem;
	}
	.welcome-intro {
		margin: 0 0 1.25rem 0;
		color: #555;
	}
	.list-nav {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
	}
	.list-nav-btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.6rem 1rem;
		border-radius: 0.5rem;
		background: var(--femtocrank-bg, #fff);
		border: 1px solid var(--femtocrank-border, #ccc);
		color: inherit;
		text-decoration: none;
		font-size: 1rem;
		transition: background 0.15s;
	}
	.list-nav-btn:hover {
		background: var(--femtocrank-hover, #f0f0f0);
	}
	.list-count {
		font-size: 0.85em;
		color: #666;
	}
	.welcome-empty,
	.welcome-loading {
		margin: 0;
		color: #666;
	}
	.welcome-error {
		margin: 0;
		color: #c00;
		font-size: 0.95rem;
	}
	.welcome-retry-btn {
		margin-left: 0.5rem;
		padding: 0.35rem 0.6rem;
		border: 1px solid #ccc;
		border-radius: 0.35rem;
		background: #fff;
		font-size: 0.9rem;
		cursor: pointer;
	}
	.welcome-retry-btn:hover {
		background: #f5f5f5;
	}
	.welcome-actions {
		margin-top: 1.5rem;
	}
</style>
