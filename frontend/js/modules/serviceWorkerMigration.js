/** Legacy cache name from the pre-Workbox service worker (frontend/sw.js). */
export const LEGACY_SW_CACHE = 'wt-cache';

/**
 * Removes caches left by the old cache-first service worker so a fresh shell can
 * load through Cloudflare Access (or any auth proxy) instead of stale HTML/JS.
 */
export async function purgeLegacyServiceWorkerCaches() {
	if (typeof caches === 'undefined') {
		return;
	}
	const keys = await caches.keys();
	await Promise.all(
		keys.filter((key) => key === LEGACY_SW_CACHE).map((key) => caches.delete(key)),
	);
}
