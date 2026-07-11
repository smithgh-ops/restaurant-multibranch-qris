import type { PageServerLoad, Actions } from './$types';
import { error, fail } from '@sveltejs/kit';
import { api } from '$lib/api/client';

export const load: PageServerLoad = async ({ params }) => {
	const { token } = params;

	if (!token || token.length !== 64) {
		throw error(400, 'Token QR tidak valid');
	}

	const res = await api.selforder.getMenu(token);
	if (res.error || !res.data) {
		throw error(404, 'Token QR tidak valid atau sudah tidak aktif. Minta staf untuk membuat QR baru.');
	}

	return {
		menu: res.data,
		token
	};
};

export const actions: Actions = {
	order: async ({ params, request }) => {
		const { token } = params;
		if (!token || token.length !== 64) return fail(400, { error: 'Token tidak valid' });

		const form = await request.formData();
		const itemsJson = String(form.get('items') ?? '[]');
		const notes = String(form.get('notes') ?? '').trim() || undefined;

		let items: { menu_item_id: number; quantity: number; notes?: string }[] = [];
		try {
			items = JSON.parse(itemsJson);
		} catch {
			return fail(400, { error: 'Data pesanan tidak valid.' });
		}

		if (!items.length) return fail(400, { error: 'Keranjang kosong.' });

		const res = await api.selforder.createOrder(token, { notes, items });
		if (res.error) return fail(400, { error: res.error });
		return { success: true, order: res.data };
	}
};
