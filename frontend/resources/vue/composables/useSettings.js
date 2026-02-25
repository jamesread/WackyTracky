import { ref, computed, watch } from 'vue';

const STORAGE_KEY = 'wackytracky_settings';

const DEFAULTS = {
	dateTimeDisplayFormat: 'short', // 'locale' | 'short' | 'iso'
	useMonospaceFont: false,
	zenMode: false,
	hideFooter: false,
};

function loadFromStorage() {
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (!raw) return { ...DEFAULTS };
		const parsed = JSON.parse(raw);
		return { ...DEFAULTS, ...parsed };
	} catch {
		return { ...DEFAULTS };
	}
}

function saveToStorage(settings) {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
	} catch {
		// ignore
	}
}

// Singleton so all components share the same state; toggles in Options immediately
// affect watchers in App.vue (e.g. monospace font, zen mode, hide footer).
const settings = ref(loadFromStorage());
watch(settings, (val) => saveToStorage(val), { deep: true });

/** Re-read settings from localStorage and update the shared ref (e.g. on app mount). */
export function syncSettingsFromStorage() {
	settings.value = loadFromStorage();
}

/** Client-side settings backed by localStorage. */
export function useSettings() {

	function updateSetting(key, value) {
		if (!(key in DEFAULTS)) return;
		settings.value = { ...settings.value, [key]: value };
	}

	const dateTimeDisplayFormat = computed({
		get: () => settings.value.dateTimeDisplayFormat ?? DEFAULTS.dateTimeDisplayFormat,
		set: (v) => {
			settings.value = { ...settings.value, dateTimeDisplayFormat: v };
		},
	});

	const useMonospaceFont = computed({
		get: () => settings.value.useMonospaceFont ?? DEFAULTS.useMonospaceFont,
		set: (v) => {
			settings.value = { ...settings.value, useMonospaceFont: v };
		},
	});

	const zenMode = computed({
		get: () => settings.value.zenMode ?? DEFAULTS.zenMode,
		set: (v) => {
			settings.value = { ...settings.value, zenMode: v };
		},
	});

	const hideFooter = computed({
		get: () => settings.value.hideFooter ?? DEFAULTS.hideFooter,
		set: (v) => {
			settings.value = { ...settings.value, hideFooter: v };
		},
	});

	/** Format a date or datetime string for display (due dates, wait until, etc.). */
	function formatDateDisplay(isoStr) {
		if (!isoStr) return '';
		const d = new Date(isoStr.includes('T') ? isoStr : isoStr + 'T12:00:00');
		if (isNaN(d.getTime())) return isoStr;
		const format = dateTimeDisplayFormat.value;

		if (format === 'iso') {
			const y = d.getFullYear();
			const m = String(d.getMonth() + 1).padStart(2, '0');
			const day = String(d.getDate()).padStart(2, '0');
			if (isoStr.includes('T')) {
				const h = String(d.getHours()).padStart(2, '0');
				const min = String(d.getMinutes()).padStart(2, '0');
				return `${y}-${m}-${day} ${h}:${min}`;
			}
			return `${y}-${m}-${day}`;
		}

		if (format === 'locale') {
			if (isoStr.includes('T')) {
				return d.toLocaleString(undefined, {
					month: 'short',
					day: 'numeric',
					year: d.getFullYear() !== new Date().getFullYear() ? 'numeric' : undefined,
					hour: 'numeric',
					minute: '2-digit',
				});
			}
			return d.toLocaleDateString(undefined, {
				month: 'short',
				day: 'numeric',
				year: d.getFullYear() !== new Date().getFullYear() ? 'numeric' : undefined,
			});
		}

		// 'short' (default): abbreviated
		if (isoStr.includes('T')) {
			const datePart = d.toLocaleDateString(undefined, {
				month: 'short',
				day: 'numeric',
				year: d.getFullYear() !== new Date().getFullYear() ? 'numeric' : undefined,
			});
			const timePart = d.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' });
			return `${datePart}, ${timePart}`;
		}
		return d.toLocaleDateString(undefined, {
			month: 'short',
			day: 'numeric',
			year: d.getFullYear() !== new Date().getFullYear() ? 'numeric' : undefined,
		});
	}

	return {
		settings,
		updateSetting,
		syncSettingsFromStorage,
		dateTimeDisplayFormat,
		useMonospaceFont,
		zenMode,
		hideFooter,
		formatDateDisplay,
		dateTimeFormatOptions: [
			{ value: 'short', label: 'Short (e.g. Jan 15, 2:30 PM)' },
			{ value: 'locale', label: 'Locale (browser default)' },
			{ value: 'iso', label: 'ISO (e.g. 2025-01-15 14:30)' },
		],
	};
}
