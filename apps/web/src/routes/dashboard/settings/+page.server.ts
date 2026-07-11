import type { PageServerLoad, Actions } from './$types';
import { redirect, fail } from '@sveltejs/kit';
import { api } from '$lib/api/client';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.user || !locals.accessToken) {
		throw redirect(302, '/');
	}

	const [orgRes, meRes] = await Promise.all([
		api.organization(locals.accessToken),
		api.me(locals.accessToken)
	]);

	return {
		organization: orgRes.data ?? null,
		me: meRes.data ?? null
	};
};

export const actions: Actions = {
	updateOrg: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi', section: 'org' });

		const form = await request.formData();
		const name = String(form.get('name') ?? '').trim() || undefined;
		const slug = String(form.get('slug') ?? '').trim() || undefined;

		if (!name && !slug) {
			return fail(400, { error: 'Minimal satu field harus diisi.', section: 'org' });
		}

		const res = await api.updateOrganization(locals.accessToken, { name, slug });
		if (res.error) {
			return fail(400, { error: res.error, section: 'org' });
		}
		return { success: true, section: 'org' };
	},

	changePassword: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi', section: 'password' });

		const form = await request.formData();
		const current_password = String(form.get('current_password') ?? '').trim();
		const new_password = String(form.get('new_password') ?? '').trim();
		const confirm_password = String(form.get('confirm_password') ?? '').trim();

		if (!current_password || !new_password) {
			return fail(400, { error: 'Semua field password wajib diisi.', section: 'password' });
		}
		if (new_password !== confirm_password) {
			return fail(400, { error: 'Password baru dan konfirmasi tidak cocok.', section: 'password' });
		}
		if (new_password.length < 6) {
			return fail(400, { error: 'Password baru minimal 6 karakter.', section: 'password' });
		}

		const res = await api.changePassword(locals.accessToken, { current_password, new_password });
		if (res.error) {
			return fail(400, { error: res.error, section: 'password' });
		}
		return { success: true, section: 'password' };
	}
};
