<template>
	<Section
		title="Welcome"
		:icon="HomeIcon"
		subtitle="Choose a list to view or add tasks."
		:padding="true"
	>
		<template #toolbar>
			<button
				type="button"
				class="inline-icon neutral"
				aria-label="Refresh"
				title="Refresh"
				:disabled="loading"
				@click="loadLists"
			>
				<HugeiconsIcon
					:icon="Refresh01Icon"
					width="1em"
					height="1em"
					:strokeWidth="iconStrokeWidth"
					aria-hidden="true"
				/>
			</button>
			<button
				type="button"
				class="inline-icon good"
				aria-label="New list"
				title="New list"
				:disabled="!isOnline"
				@click="showCreateDialog = true"
			>
				<HugeiconsIcon
					:icon="PlusSignIcon"
					width="1em"
					height="1em"
					:strokeWidth="iconStrokeWidth"
					aria-hidden="true"
				/>
			</button>
		</template>

		<p v-if="loadError" class="inline-notification error" role="alert">
			{{ loadError }}
			<button type="button" class="neutral small" @click="loadLists">Try again</button>
		</p>
		<p v-else-if="loading && !lists.length" class="muted">Loading…</p>
		<p v-else-if="!loading && !lists.length" class="inline-notification note">No lists yet.</p>
		<Navigation v-else ref="navRef">
			<NavigationGrid />
		</Navigation>

		<CreateListDialog
			v-model="showCreateDialog"
			@created="onListCreated"
		/>
	</Section>
</template>

<script setup>
	import { ref, onMounted, watch, inject, nextTick } from 'vue';
	import { useRouter } from 'vue-router';
	import Section from 'picocrank/vue/components/Section.vue';
	import Navigation from 'picocrank/vue/components/Navigation.vue';
	import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue';
	import { HugeiconsIcon } from '@hugeicons/vue';
	import { HomeIcon, Folder01Icon, PlusSignIcon, Refresh01Icon } from '@hugeicons/core-free-icons';
	import CreateListDialog from '../components/CreateListDialog.vue';
	import { getCachedInbox, INBOX_LIST_ID } from '../../../js/modules/offlineStorage.js';
	import { describeApiError } from '../../../js/modules/apiError.js';

	const iconStrokeWidth = 2.5;
	const router = useRouter();
	const lists = ref([]);
	const loading = ref(true);
	const loadError = ref(null);
	const showCreateDialog = ref(false);
	const navRef = ref(null);
	const refreshTrigger = inject('refreshTrigger', null);
	const showToast = inject('showToast', () => {});
	const isOnline = inject('isOnline', ref(true));

	async function syncNavGrid() {
		await nextTick();
		const nav = navRef.value;
		if (!nav) return;
		nav.clearNavigationLinks();
		for (const list of lists.value) {
			nav.addCallback(list.title, () => router.push('/lists/' + list.id), {
				icon: Folder01Icon,
				name: 'list-' + list.id,
				count: list.countItems,
			});
		}
	}

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
			await syncNavGrid();
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
			const message = describeApiError(e, 'Could not load lists');
			loadError.value = message;
			showToast(message, 'error');
		} finally {
			loading.value = false;
			await syncNavGrid();
		}
	}

	function onListCreated() {
		loadLists();
		if (refreshTrigger) refreshTrigger.value++;
	}

	watch(isOnline, loadLists);
	watch(lists, () => { syncNavGrid(); });
	onMounted(loadLists);
</script>
