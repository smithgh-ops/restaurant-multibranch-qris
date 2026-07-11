import type { PageServerLoad, Actions } from './$types';
import { redirect, fail } from '@sveltejs/kit';
import { api, API_BASE_URL } from '$lib/api/client';

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
		const imageFile = form.get('image') as File | null;

		if (!category_id || !name || !base_price) {
			return fail(400, { error: 'Kategori, nama, dan harga wajib diisi.' });
		}

		// Step 1: create the item (no image yet).
		const res = await api.menu.items.create(locals.accessToken, {
			category_id,
			name,
			description,
			base_price
		});
		if (res.error) return fail(400, { error: res.error });

		// Step 2: upload image if provided.
		if (imageFile && imageFile.size > 0) {
			const itemId = res.data!.id;
			const upstream = new FormData();
			upstream.append('image', imageFile, imageFile.name);
			try {
				const imgRes = await fetch(`${API_BASE_URL}/api/v1/menu/items/${itemId}/image`, {
					method: 'POST',
					headers: { Authorization: 'Bearer ' + locals.accessToken },
					body: upstream
				});
				if (!imgRes.ok) {
					// Item was created but image upload failed — still report success with a warning.
					return { success: true, action: 'createItem', warning: 'Item berhasil dibuat, tapi gambar gagal diunggah.' };
				}
			} catch {
				return { success: true, action: 'createItem', warning: 'Item berhasil dibuat, tapi gambar gagal diunggah.' };
			}
		}

		return { success: true, action: 'createItem' };
	},

	updateItem: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const id = Number(form.get('id'));
		const name = String(form.get('name') ?? '').trim() || undefined;
		const base_price = String(form.get('base_price') ?? '').trim() || undefined;
		const description = String(form.get('description') ?? '').trim() || undefined;
		const image_url = String(form.get('image_url') ?? '').trim() || undefined;
		const isActiveRaw = form.get('is_active');
		const is_active = isActiveRaw !== null ? isActiveRaw === 'true' : undefined;

		if (!id) return fail(400, { error: 'ID item tidak valid.' });

		const res = await api.menu.items.update(locals.accessToken, id, {
			name,
			base_price,
			description,
			image_url,
			is_active
		});
		if (res.error) return fail(400, { error: res.error });
		return { success: true, action: 'updateItem' };
	},

	uploadImage: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const id = Number(form.get('id'));
		const imageFile = form.get('image') as File | null;

		if (!id) return fail(400, { error: 'ID item tidak valid.' });
		if (!imageFile || imageFile.size === 0) return fail(400, { error: 'File gambar tidak ditemukan.' });

		// Forward the multipart upload directly to the API.
		const upstream = new FormData();
		upstream.append('image', imageFile, imageFile.name);
		try {
			const res = await fetch(`${API_BASE_URL}/api/v1/menu/items/${id}/image`, {
				method: 'POST',
				headers: { Authorization: 'Bearer ' + locals.accessToken },
				body: upstream
			});
			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				return fail(res.status, { error: body.error ?? `HTTP ${res.status}` });
			}
		} catch {
			return fail(500, { error: 'Gagal mengunggah gambar.' });
		}
		return { success: true, action: 'uploadImage' };
	},

	deleteImage: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const id = Number(form.get('id'));
		if (!id) return fail(400, { error: 'ID item tidak valid.' });

		const res = await api.menu.items.deleteImage(locals.accessToken, id);
		if (res.error) return fail(400, { error: res.error });
		return { success: true, action: 'deleteImage' };
	}
};
