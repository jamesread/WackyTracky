import assert from 'node:assert/strict'
import { test } from 'node:test'
import { clearSSOOptOutCookie, hasSSOOptOut, SSO_OPT_OUT_COOKIE } from '../js/modules/ssoOptOut.js'

test('detects the opt-out cookie among others', () => {
	assert.equal(hasSSOOptOut(`foo=bar; ${SSO_OPT_OUT_COOKIE}=1; baz=qux`), true)
	assert.equal(hasSSOOptOut(`${SSO_OPT_OUT_COOKIE}=1`), true)
})

test('ignores a missing, empty or similarly named cookie', () => {
	assert.equal(hasSSOOptOut(''), false)
	assert.equal(hasSSOOptOut(undefined), false)
	assert.equal(hasSSOOptOut('foo=bar'), false)
	assert.equal(hasSSOOptOut(`${SSO_OPT_OUT_COOKIE}=`), false)
	assert.equal(hasSSOOptOut(`x${SSO_OPT_OUT_COOKIE}=1`), false)
})

test('builds a cookie string that expires the opt-out', () => {
	assert.equal(clearSSOOptOutCookie(false), `${SSO_OPT_OUT_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`)
	assert.match(clearSSOOptOutCookie(true), /; Secure$/)
})
