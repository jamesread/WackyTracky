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
				<h2 id="create-list-title" class="create-list-dialog-title">New list</h2>
				<form @submit.prevent="submit" class="create-list-dialog-form">
					<label for="create-list-name">List name</label>
					<input
						id="create-list-name"
						v-model="title"
						type="text"
						autofocus
						required
						placeholder="Enter list name"
						class="create-list-dialog-input"
					/>
					<div class="create-list-dialog-actions">
						<button type="button" class="create-list-dialog-btn" @click="cancel">
							Cancel
						</button>
						<button type="submit" class="create-list-dialog-btn create-list-dialog-btn-primary">
							Create
						</button>
					</div>
				</form>
				<p v-if="error" class="create-list-dialog-error">{{ error }}</p>
			</div>
		</div>
	</Teleport>
</template>

<script setup>
import { ref, watch } from 'vue';

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
	background: var(--femtocrank-bg, #fff);
	border-radius: 8px;
	box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
	min-width: 320px;
	max-width: 90vw;
	padding: 1.25rem;
}

.create-list-dialog-title {
	margin: 0 0 1rem;
	font-size: 1.25rem;
}

.create-list-dialog-form {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.create-list-dialog-form label {
	display: block;
	font-weight: 500;
}

.create-list-dialog-input {
	width: 100%;
	padding: 0.5rem 0.75rem;
	border: 1px solid #ccc;
	border-radius: 4px;
	box-sizing: border-box;
}

.create-list-dialog-actions {
	display: flex;
	gap: 0.5rem;
	justify-content: flex-end;
	margin-top: 0.25rem;
}

.create-list-dialog-btn {
	padding: 0.5rem 1rem;
	border-radius: 4px;
	border: 1px solid #ccc;
	background: var(--femtocrank-bg, #fff);
	cursor: pointer;
}

.create-list-dialog-btn-primary {
	background: var(--femtocrank-primary, #2dae82);
	color: #fff;
	border-color: var(--femtocrank-primary, #2dae82);
}

.create-list-dialog-error {
	margin: 0.75rem 0 0;
	color: #c00;
	font-size: 0.9rem;
}
</style>
