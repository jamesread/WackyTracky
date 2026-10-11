import { reactive, readonly } from 'vue'
import { Code, ConnectError } from '@connectrpc/connect'

const state = reactive({
	loaded: false,
	loading: false,
	error: /** @type {string | null} */ (null),
	/** When true, GetStatus is unavailable (proto/codegen not ready); app runs without login gate. */
	compatNoAuth: false,
	status: /** @type {Record<string, unknown> | null} */ (null),
})

/** @type {Promise<Record<string, unknown>> | null} */
let inflight = null

export function useAppStatus() {
	return readonly(state)
}

/**
 * @param {Record<string, unknown> | null} st
 */
export function applyStatusToWindow(st) {
	if (typeof window === 'undefined') return
	window.userRbacPermissions = /** @type {string[]} */ (st?.rbacPermissions || [])
	window.userRbacIsSuperuser = Boolean(st?.rbacIsSuperuser)
}

export function invalidateAppStatus() {
	state.loaded = false
	state.status = null
	state.compatNoAuth = false
	inflight = null
}

function compatStatus() {
	return {
		isLoggedIn: true,
		rbacIsSuperuser: true,
		rbacPermissions: [],
		username: '',
	}
}

function isMissingGetStatus(err) {
	if (typeof window !== 'undefined' && typeof window.client?.getStatus !== 'function') {
		return true
	}
	if (err instanceof ConnectError && err.code === Code.Unimplemented) {
		return true
	}
	const msg = err instanceof Error ? err.message : String(err ?? '')
	return /unimplemented|not found|is not a function/i.test(msg)
}

/**
 * @param {{ force?: boolean }} [opts]
 * @returns {Promise<Record<string, unknown>>}
 */
export async function fetchAppStatus(opts) {
	if (!opts?.force && state.loaded && state.status) {
		return state.status
	}
	if (!opts?.force && inflight) {
		return inflight
	}

	state.loading = true
	state.error = null

	if (typeof window === 'undefined' || typeof window.client?.getStatus !== 'function') {
		const st = compatStatus()
		state.status = st
		state.loaded = true
		state.compatNoAuth = true
		state.loading = false
		applyStatusToWindow(st)
		return st
	}

	inflight = window.client
		.getStatus({})
		.then((st) => {
			state.status = st
			state.loaded = true
			state.compatNoAuth = false
			applyStatusToWindow(st)
			return st
		})
		.catch((err) => {
			if (isMissingGetStatus(err)) {
				const st = compatStatus()
				state.status = st
				state.loaded = true
				state.compatNoAuth = true
				applyStatusToWindow(st)
				return st
			}
			state.error = err instanceof Error ? err.message : 'Status failed'
			throw err
		})
		.finally(() => {
			state.loading = false
			inflight = null
		})

	return inflight
}
