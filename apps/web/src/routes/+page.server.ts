import type { Actions } from './$types';
import { redirect } from '@sveltejs/kit';
import { api } from '$lib/api/client';

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const form = await request.formData();
		const email = String(form.get('email') ?? '');
		const password = String(form.get('password') ?? '');

		if (!email || !password) {
			return { success: false, error: 'Email dan password wajib diisi.' };
		}

		const res = await api.login(email, password);
		if (res.error || !res.data) {
			return {
				success: false,
				error: res.error ?? 'Login gagal, silakan coba lagi.'
			};
		}

		const { access_token, expires_in, refresh_token } = res.data;

		cookies.set('access_token', access_token, {
			httpOnly: true,
			path: '/',
			maxAge: expires_in ?? 900,
			sameSite: 'lax',
			secure: false // set true in production behind HTTPS
		});

		if (refresh_token) {
			cookies.set('refresh_token', refresh_token, {
				httpOnly: true,
				path: '/',
				maxAge: 30 * 24 * 60 * 60,
				sameSite: 'lax',
				secure: false
			});
		}

		throw redirect(302, '/dashboard');
	}
};
