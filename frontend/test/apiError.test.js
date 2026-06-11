import { test } from 'node:test';
import assert from 'node:assert/strict';
import { Code, ConnectError } from '@connectrpc/connect';

import { describeApiError } from '../js/modules/apiError.js';

// connect-web builds errors for HTTP failures as `new ConnectError("HTTP <status>", codeFromHttpStatus(status))`.
function httpError(status, code) {
	return new ConnectError(`HTTP ${status}`, code);
}

test('maps HTTP 403 from a proxy to a forbidden/re-auth hint', () => {
	const msg = describeApiError(httpError(403, Code.PermissionDenied), 'Could not load lists');
	assert.match(msg, /Could not load lists:/);
	assert.match(msg, /403/);
	assert.match(msg, /forbidden/i);
});

test('maps HTTP 404 to an endpoint-not-found hint', () => {
	const msg = describeApiError(httpError(404, Code.Unimplemented));
	assert.match(msg, /404/);
	assert.match(msg, /not found/i);
});

test('maps HTTP 401 to an authentication-required hint', () => {
	const msg = describeApiError(httpError(401, Code.Unauthenticated));
	assert.match(msg, /401/);
	assert.match(msg, /[Aa]uthentication/);
});

test('maps HTTP 502 to a bad-gateway/proxy hint', () => {
	const msg = describeApiError(httpError(502, Code.Unavailable));
	assert.match(msg, /502/);
	assert.match(msg, /gateway|proxy/i);
});

test('detects an HTML login/error page parsed as JSON', () => {
	const err = new ConnectError('Unexpected token < in JSON at position 0', Code.Internal);
	const msg = describeApiError(err);
	assert.match(msg, /unexpected response|proxy|sign in/i);
});

test('falls back to a Connect code hint when no HTTP status is present', () => {
	const err = new ConnectError('connection reset', Code.Unavailable);
	const msg = describeApiError(err);
	assert.match(msg, /unavailable/i);
});

test('describes a network-level TypeError without an HTTP status', () => {
	const msg = describeApiError(new TypeError('Failed to fetch'));
	assert.match(msg, /network error/i);
	assert.doesNotMatch(msg, /HTTP/);
});

test('prefixes with context when provided', () => {
	const msg = describeApiError(httpError(403, Code.PermissionDenied), 'Search failed');
	assert.ok(msg.startsWith('Search failed: '));
});

test('handles unknown/plain errors gracefully', () => {
	assert.equal(typeof describeApiError(new Error('boom')), 'string');
	assert.match(describeApiError(new Error('boom')), /boom/);
});
