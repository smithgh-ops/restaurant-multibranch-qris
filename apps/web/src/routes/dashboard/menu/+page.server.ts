import type { PageServerLoad, Actions } from './$types';
import { redirect, fail } from '@sveltejs/kit';
import { api } from '$lib/api/client';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.user || !locals.accessToken) {
		throw redirect(302, '/');
	}

	const [categoriesRes, itemsRes] = await Promise.all([
		api.menu.categories.list(locals.accessToken),
		api.menu.items.list(locals.accessToken)
	]);

	return {
		categories: categoriesRes.data?.data ?? [],
		items: itemsRes.data?.data ?? [],
		error: categoriesRes.error ?? itemsRes.error
	};
};

export const actions: Actions = {
	createCategory: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const name = String(form.get('name') ?? '').trim();
		const description = String(form.get('description') ?? '').trim() || undefined;

		if (!name) return fail(400, { error: 'Nama kategori wajib diisi.' });

		const res = await api.menu.categories.create(locals.accessToken, { name, description });
		if (res.error) return fail(400, { error: res.error });
		return { success: true, action: 'createCategory' };
	},

	createItem: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const category_id = Number(form.get('category_id'));
		const name = String(form.get('name') ?? '').trim();
		const description = String(form.get('description') ?? '').trim() || undefined;
		const base_price = String(form.get('base_price') ?? '').trim();

		if (!category_id || !name || !base_price) {
			return fail(400, { error: 'Kategori, nama, dan harga wajib diisi.' });
		}

		const res = await api.menu.items.create(locals.accessToken, {
			category_id,
			name,
			description,
			base_price
		});
		if (res.error) return fail(400, { error: res.error });
		return { success: true, action: 'createItem' };
	},

	updateItem: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const id = Number(form.get('id'));
		const name = String(form.get('name') ?? '').trim() || undefined;
		const base_price = String(form.get('base_price') ?? '').trim() || undefined;
		const description = String(form.get('description') ?? '').trim() || undefined;
		const isActiveRaw = form.get('is_active');
		const is_active = isActiveRaw !== null ? isActiveRaw === 'true' : undefined;

		if (!id) return fail(400, { error: 'ID item tidak valid.' });

		const res = await api.menu.items.update(locals.accessToken, id, {
			name,
			base_price,
			description,
			is_active
		});
		if (res.error) return fail(400, { error: res.error });
		return { success: true, action: 'updateItem' };
	}
};
