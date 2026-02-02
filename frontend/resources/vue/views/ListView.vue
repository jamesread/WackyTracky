<template>

	<section class = "transparent">
		<div v-if="!items.length" style = "text-align: center; padding: 1rem;">
				<p>This list is empty.</p>
		</div>
		<div v-else>
			<ul>
				<li v-for="item in items" :key="item.id">
					{{ item.content }}
				</li>
			</ul>
		</div>
	</section>
</template>

<script setup>
	import { onMounted } from 'vue';
	import { ref } from 'vue';
	const items = ref([]);

	const props = defineProps({
		listId: {
			type: String,
			default: null,
		},
	});

	async function getLists() {
		const ret = await window.client.listTasks({
            parentType: 'list',
			parentId: props.listId,
		});

		items.value = ret.tasks
	}
	
	onMounted(() => {
        getLists();
	})
</script>

<style scoped>
    section {
		padding: 0;
		margin: 0;
	}

	li {
		background-color: white;
		padding: 0.5rem;
		border-radius: 0.5rem;
		margin-bottom: 1.2rem;
		box-shadow: 0 0 .5em #9a9a9a;
	}

	ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

</style>
