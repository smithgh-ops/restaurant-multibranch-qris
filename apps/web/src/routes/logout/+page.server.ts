import type { Actions } from './$types';
import { redirect } from '@sveltejs/kit';
import { api } from '$lib/api/client';

export const actions: Actions = {
	default: async ({ cookies, locals }) => {
		const refreshToken = cookies.get('refresh_token');
		if (locals.accessToken) {
			await api.logout(locals.accessToken, refreshToken);
		}
		cookies.delete('access_token', { path: '/' });
		cookies.delete('refresh_token', { path: '/' });
		throw redirect(302, '/');
	}
};
