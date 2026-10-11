import { describe, it } from 'node:test'
import assert from 'node:assert/strict'
import { showUpdateLinkFromInit } from '../js/modules/appInit.js'

describe('showUpdateLinkFromInit', () => {
	it('hides when version number policy is off', () => {
		assert.equal(
			showUpdateLinkFromInit({
				showVersionNumber: false,
				showNewVersions: true,
				availableVersion: '1.2.3',
			}),
			false,
		)
	})

	it('hides when showNewVersions is false', () => {
		assert.equal(
			showUpdateLinkFromInit({
				showVersionNumber: true,
				showNewVersions: false,
				availableVersion: '1.2.3',
			}),
			false,
		)
	})

	it('hides sentinels', () => {
		assert.equal(
			showUpdateLinkFromInit({
				showVersionNumber: true,
				showNewVersions: true,
				availableVersion: 'none',
			}),
			false,
		)
		assert.equal(
			showUpdateLinkFromInit({
				showVersionNumber: true,
				showNewVersions: true,
				availableVersion: 'you-are-using-a-dev-build',
			}),
			false,
		)
	})

	it('shows a real newer version', () => {
		assert.equal(
			showUpdateLinkFromInit({
				showVersionNumber: true,
				showNewVersions: true,
				availableVersion: '2.0.0',
			}),
			true,
		)
	})
})
