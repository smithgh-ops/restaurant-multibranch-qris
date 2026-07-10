import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';
import { api } from '$lib/api/client';

// Routes that do NOT require authentication
const PUBLIC_PATHS = new Set(['/', '/login']);

export const handle: Handle = async ({ event, resolve }) => {
	const path = event.url.pathname;

	// Read tokens from httpOnly cookies
	const accessToken = event.cookies.get('access_token');
	const refreshToken = event.cookies.get('refresh_token');

	if (accessToken) {
		// Try to populate locals from the existing access token
		const res = await api.me(accessToken);
		if (res.data) {
			event.locals.user = {
				id: res.data.id,
				organizationId: res.data.organization_id,
				name: res.data.name,
				email: res.data.email
			};
			event.locals.accessToken = accessToken;
		} else if (refreshToken) {
			// Access token may be invalid/expired — try to refresh
			const refreshRes = await api.refresh(refreshToken);
			if (refreshRes.data) {
				const newAccess = refreshRes.data.access_token;
				const newRefresh = refreshRes.data.refresh_token;

				event.cookies.set('access_token', newAccess, {
					httpOnly: true,
					path: '/',
					maxAge: (refreshRes.data.expires_in ?? 900),
					sameSite: 'lax',
					secure: false // set to true in production behind HTTPS
				});
				if (newRefresh) {
					event.cookies.set('refresh_token', newRefresh, {
						httpOnly: true,
						path: '/',
						maxAge: 30 * 24 * 60 * 60,
						sameSite: 'lax',
						secure: false
					});
				}

				const meRes = await api.me(newAccess);
				if (meRes.data) {
					event.locals.user = {
						id: meRes.data.id,
						organizationId: meRes.data.organization_id,
						name: meRes.data.name,
						email: meRes.data.email
					};
					event.locals.accessToken = newAccess;
				}
			}
		}
	} else if (refreshToken) {
		// No access token but there is a refresh token — refresh silently
		const refreshRes = await api.refresh(refreshToken);
		if (refreshRes.data) {
			const newAccess = refreshRes.data.access_token;
			const newRefresh = refreshRes.data.refresh_token;

			event.cookies.set('access_token', newAccess, {
				httpOnly: true,
				path: '/',
				maxAge: refreshRes.data.expires_in ?? 900,
				sameSite: 'lax',
				secure: false
			});
			if (newRefresh) {
				event.cookies.set('refresh_token', newRefresh, {
					httpOnly: true,
					path: '/',
					maxAge: 30 * 24 * 60 * 60,
					sameSite: 'lax',
					secure: false
				});
			}

			const meRes = await api.me(newAccess);
			if (meRes.data) {
				event.locals.user = {
					id: meRes.data.id,
					organizationId: meRes.data.organization_id,
					name: meRes.data.name,
					email: meRes.data.email
				};
				event.locals.accessToken = newAccess;
			}
		}
	}

	// Protect non-public routes
	if (!PUBLIC_PATHS.has(path) && !path.startsWith('/dashboard') === false) {
		// path is a dashboard route
	}
	if (path.startsWith('/dashboard') && !event.locals.user) {
		throw redirect(302, '/');
	}
	// Redirect already-logged-in users away from login
	if ((path === '/' || path === '/login') && event.locals.user) {
		throw redirect(302, '/dashboard');
	}

	return resolve(event);
};
