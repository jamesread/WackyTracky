import { test } from 'node:test';
import assert from 'node:assert/strict';

import { LEGACY_SW_CACHE, purgeLegacyServiceWorkerCaches } from '../js/modules/serviceWorkerMigration.js';

test('purgeLegacyServiceWorkerCaches deletes only the legacy wt-cache', async () => {
	const deleted = [];
	const fakeCaches = {
		keys: async () => ['wt-cache', 'pages-cache', 'other'],
		delete: async (key) => {
			deleted.push(key);
			return true;
		},
	};

	globalThis.caches = fakeCaches;
	await purgeLegacyServiceWorkerCaches();
	delete globalThis.caches;

	assert.deepEqual(deleted, [LEGACY_SW_CACHE]);
});

test('purgeLegacyServiceWorkerCaches is a no-op when caches is unavailable', async () => {
	delete globalThis.caches;
	await purgeLegacyServiceWorkerCaches();
});
