<template>
	<Teleport to="body">
		<div
			v-if="modelValue"
			class="create-list-dialog-overlay"
			role="dialog"
			aria-modal="true"
			aria-labelledby="create-list-title"
			@click.self="cancel"
		>
			<div class="create-list-dialog">
				<h2 id="create-list-title">New list</h2>
				<FormLayout @submit.prevent="submit">
					<FormField label="List name" for="create-list-name">
						<input
							id="create-list-name"
							v-model="title"
							type="text"
							autofocus
							required
							placeholder="Enter list name"
						/>
					</FormField>
					<p v-if="error" class="inline-notification error">{{ error }}</p>
					<template #actions>
						<button type="button" class="neutral" @click="cancel">Cancel</button>
						<button type="submit" class="good" :disabled="!title.trim()">Create</button>
					</template>
				</FormLayout>
			</div>
		</div>
	</Teleport>
</template>

<script setup>
import { ref, watch } from 'vue';
import FormField from 'picocrank/vue/components/FormField.vue';
import FormLayout from 'picocrank/vue/components/FormLayout.vue';

const props = defineProps({
	modelValue: { type: Boolean, default: false },
});

const emit = defineEmits(['update:modelValue', 'created']);

const title = ref('');
const error = ref('');

watch(() => props.modelValue, (open) => {
	if (open) {
		title.value = '';
		error.value = '';
	}
});

function cancel() {
	emit('update:modelValue', false);
}

async function submit() {
	error.value = '';
	if (!title.value.trim()) return;
	if (!window.client) {
		error.value = 'Not connected';
		return;
	}
	try {
		await window.client.createList({ title: title.value.trim() });
		emit('created');
		emit('update:modelValue', false);
	} catch (e) {
		error.value = e?.message || 'Failed to create list';
	}
}
</script>

<style scoped>
.create-list-dialog-overlay {
	position: fixed;
	inset: 0;
	background: rgba(0, 0, 0, 0.4);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1000;
}

.create-list-dialog {
	background: var(--section-bg-color, #fff);
	border-radius: 0.8em;
	box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
	min-width: 320px;
	max-width: 90vw;
	padding: 1.25rem;
}

.create-list-dialog h2 {
	margin: 0 0 1rem;
	font-size: 1.25rem;
}
</style>
