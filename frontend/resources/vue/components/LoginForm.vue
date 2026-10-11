<script setup>
import { ref } from 'vue'
import Login from 'picocrank/vue/components/Login.vue'
import Section from 'picocrank/vue/components/Section.vue'

const emit = defineEmits(['login-success'])
const loginRef = ref(null)
const logoUrl = new URL('../../images/logos/wacky-tracky.png', import.meta.url).href

async function onLocalLogin(payload) {
	try {
		const res = await window.client.loginWithUsernameAndPassword({
			username: payload.username,
			password: payload.password,
		})
		if (res.standardResponse && !res.standardResponse.success) {
			loginRef.value?.setLocalLoginError(res.standardResponse.message || 'Login failed')
			return
		}
		loginRef.value?.resetLocalForm()
		emit('login-success')
	} catch {
		loginRef.value?.setLocalLoginError('Login failed')
	}
}
</script>

<template>
	<Section subtitle="Sign in with your username and password" :padding="true">
		<template #title>
			<span class="section-title-with-icon">
				<img :src="logoUrl" alt="" class="section-title-logo" aria-hidden="true" />
				<span>Login to WackyTracky</span>
			</span>
		</template>
		<Login
			ref="loginRef"
			local-only
			@local-login="onLocalLogin"
		/>
	</Section>
</template>

<style scoped>
.section-title-with-icon {
	display: inline-flex;
	align-items: center;
	gap: 0.45em;
	vertical-align: middle;
}

.section-title-logo {
	width: 1.1em;
	height: 1.1em;
	flex-shrink: 0;
}
</style>
