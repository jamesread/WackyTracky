/**
 * Turns a failed API call (Connect RPC error, network failure, or an
 * unexpected response from an authenticating proxy) into a short, descriptive
 * message for the user. The app is often deployed behind proxies that return
 * 401/403/404 or an HTML login page instead of the API, and a generic
 * "please try again" hides which of those happened.
 */

import { Code, ConnectError } from '@connectrpc/connect';

const HTTP_STATUS_HINTS = {
	401: 'Authentication required (HTTP 401). A proxy in front of the app may need you to sign in again.',
	403: 'Access forbidden (HTTP 403). A proxy or the server is blocking API requests — you may need to re-authenticate.',
	404: 'API endpoint not found (HTTP 404). The backend may be unreachable, or a proxy is not forwarding /api requests.',
	408: 'The request timed out (HTTP 408). The server or a proxy took too long to respond.',
	429: 'Too many requests (HTTP 429). The server or a proxy is rate-limiting; wait a moment and retry.',
	500: 'The server hit an internal error (HTTP 500).',
	502: 'Bad gateway (HTTP 502). A proxy could not reach the backend.',
	503: 'The server is unavailable (HTTP 503). It may be down, restarting, or blocked by a proxy.',
	504: 'Gateway timeout (HTTP 504). A proxy did not get a response from the backend in time.',
};

const CODE_HINTS = {
	[Code.Unauthenticated]: 'Authentication required. A proxy in front of the app may need you to sign in again.',
	[Code.PermissionDenied]: 'Access forbidden. A proxy or the server is blocking API requests.',
	[Code.Unimplemented]: 'API endpoint not found. The backend may be unreachable or misconfigured.',
	[Code.Unavailable]: 'The server is unavailable. It may be down, restarting, or blocked by a proxy.',
	[Code.DeadlineExceeded]: 'The request timed out before the server responded.',
};

function httpStatusFromMessage(message) {
	const match = /HTTP (\d{3})/.exec(message ?? '');
	return match ? Number(match[1]) : null;
}

// A proxy redirect/login page or error page is served as HTML; the JSON parser
// then fails, which is a strong signal the response was not the API at all.
function looksLikeNonApiResponse(message) {
	return /unexpected token|in JSON|not valid JSON|unexpected end of|<!doctype|<html/i.test(message ?? '');
}

function apiErrorDetail(err) {
	if (err == null) {
		return 'Unknown error.';
	}

	// fetch() rejects with a TypeError for network-level failures (offline, DNS,
	// connection refused, CORS) before any HTTP status exists.
	if (err instanceof TypeError) {
		return 'Could not reach the server (network error). Check your connection and that the server is running.';
	}

	const connectError = err instanceof ConnectError ? err : null;
	const rawMessage = connectError ? connectError.rawMessage : (err.message ?? String(err));

	const status = httpStatusFromMessage(rawMessage) ?? httpStatusFromMessage(connectError?.message);
	if (status && HTTP_STATUS_HINTS[status]) {
		return HTTP_STATUS_HINTS[status];
	}

	if (looksLikeNonApiResponse(rawMessage)) {
		return 'The server returned an unexpected response (likely an HTML login or error page from a proxy, not the API). You may need to sign in again.';
	}

	if (connectError && CODE_HINTS[connectError.code]) {
		return CODE_HINTS[connectError.code];
	}

	if (status) {
		return `The server returned HTTP ${status}.`;
	}

	return rawMessage || 'Unknown error.';
}

/**
 * @param {unknown} err The thrown error from an API call.
 * @param {string} [context] Optional prefix describing the action (e.g. "Could not load lists").
 * @returns {string} A descriptive, user-facing message.
 */
export function describeApiError(err, context) {
	const detail = apiErrorDetail(err);
	return context ? `${context}: ${detail}` : detail;
}
