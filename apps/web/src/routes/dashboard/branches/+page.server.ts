import type { PageServerLoad, Actions } from './$types';
import { redirect, fail } from '@sveltejs/kit';
import { api } from '$lib/api/client';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.user || !locals.accessToken) {
		throw redirect(302, '/');
	}

	const res = await api.branches.list(locals.accessToken);
	return {
		branches: res.data?.data ?? [],
		error: res.error
	};
};

export const actions: Actions = {
	create: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const name = String(form.get('name') ?? '').trim();
		const slug = String(form.get('slug') ?? '').trim();
		const address = String(form.get('address') ?? '').trim() || undefined;
		const phone = String(form.get('phone') ?? '').trim() || undefined;

		if (!name || !slug) {
			return fail(400, { error: 'Nama dan slug wajib diisi.' });
		}

		const res = await api.branches.create(locals.accessToken, { name, slug, address, phone });
		if (res.error) {
			return fail(400, { error: res.error });
		}
		return { success: true };
	},

	update: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const id = Number(form.get('id'));
		const name = String(form.get('name') ?? '').trim() || undefined;
		const slug = String(form.get('slug') ?? '').trim() || undefined;
		const address = String(form.get('address') ?? '').trim() || undefined;
		const phone = String(form.get('phone') ?? '').trim() || undefined;
		const isActiveRaw = form.get('is_active');
		const is_active = isActiveRaw !== null ? isActiveRaw === 'true' : undefined;

		if (!id) return fail(400, { error: 'ID cabang tidak valid.' });

		const res = await api.branches.update(locals.accessToken, id, { name, slug, address, phone, is_active });
		if (res.error) {
			return fail(400, { error: res.error });
		}
		return { success: true };
	}
};
