// Set by the server on logout from a proxy JWT session (eg Cloudflare Access) so the
// JWT is ignored and a local account can sign in. Clearing it signs in with SSO again.
// Must match auth.SSOOptOutCookieName in the service.
export const SSO_OPT_OUT_COOKIE = 'wacky-tracky-sso-off'

export function hasSSOOptOut(cookieString) {
	return String(cookieString ?? '')
		.split(';')
		.map((part) => part.trim())
		.some((part) => part.startsWith(`${SSO_OPT_OUT_COOKIE}=`) && part.length > SSO_OPT_OUT_COOKIE.length + 1)
}

export function clearSSOOptOutCookie(secure) {
	return `${SSO_OPT_OUT_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax${secure ? '; Secure' : ''}`
}
