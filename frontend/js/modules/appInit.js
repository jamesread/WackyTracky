import { reactive, readonly } from 'vue'

const state = reactive({
	loaded: false,
	loading: false,
	error: /** @type {string | null} */ (null),
	init: /** @type {Record<string, unknown> | null} */ (null),
})

/** @type {Promise<Record<string, unknown>> | null} */
let inflight = null

export function useAppInit() {
	return readonly(state)
}

/**
 * @param {{ force?: boolean }} [opts]
 * @returns {Promise<Record<string, unknown>>}
 */
export async function fetchAppInit(opts) {
	if (!opts?.force && state.loaded && state.init) {
		return state.init
	}
	if (!opts?.force && inflight) {
		return inflight
	}

	state.loading = true
	state.error = null

	if (typeof window === 'undefined' || typeof window.client?.init !== 'function') {
		const fallback = {
			showFooter: true,
			showNewVersions: false,
			availableVersion: '',
			currentVersion: '',
			pageTitle: 'WackyTracky',
			showVersionNumber: true,
			wallpaper: 'wallpaper.jpg',
		}
		state.init = fallback
		state.loaded = true
		state.loading = false
		if (typeof window !== 'undefined') {
			window.initResponse = fallback
		}
		return fallback
	}

	inflight = window.client
		.init({})
		.then((res) => {
			state.init = res
			state.loaded = true
			if (typeof window !== 'undefined') {
				window.initResponse = res
			}
			return res
		})
		.catch((err) => {
			state.error = err instanceof Error ? err.message : 'Init failed'
			throw err
		})
		.finally(() => {
			state.loading = false
			inflight = null
		})

	return inflight
}

/**
 * Whether the update chip should render for an Init-like object.
 * @param {Record<string, unknown> | null | undefined} init
 */
export function showUpdateLinkFromInit(init) {
	if (!init?.showVersionNumber || !init?.showNewVersions) return false
	const v = String(init.availableVersion || '').trim()
	if (!v || v === 'none' || v === '?' || v.startsWith('you-are-using')) return false
	return true
}
