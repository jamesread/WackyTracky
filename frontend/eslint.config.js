import js from '@eslint/js';
import pluginVue from 'eslint-plugin-vue';
import globals from 'globals';

export default [
	{
		ignores: ['dist/**', 'js/gen/**', 'tmp/**', 'public/**', 'node_modules/**'],
	},
	js.configs.recommended,
	...pluginVue.configs['flat/essential'],
	{
		languageOptions: {
			ecmaVersion: 'latest',
			sourceType: 'module',
			globals: globals.browser,
		},
		rules: {
			'no-console': 'off',
			// localStorage access can throw (private windows, blocked storage) and is best-effort.
			'no-empty': ['error', { allowEmptyCatch: true }],
		},
	},
	{
		// Route views are named after their route, e.g. "Options".
		files: ['resources/vue/views/**/*.vue'],
		rules: {
			'vue/multi-word-component-names': 'off',
		},
	},
	{
		files: ['test/**/*.js', 'vite.config.js', 'buildid.js'],
		languageOptions: {
			globals: globals.node,
		},
	},
];
