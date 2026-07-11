import type { PageServerLoad, Actions } from './$types';
import { redirect, fail } from '@sveltejs/kit';
import { api } from '$lib/api/client';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.user || !locals.accessToken) {
		throw redirect(302, '/');
	}

	const [branchesRes, categoriesRes, itemsRes] = await Promise.all([
		api.branches.list(locals.accessToken),
		api.menu.categories.list(locals.accessToken),
		api.menu.items.list(locals.accessToken, { active: true })
	]);

	return {
		branches: branchesRes.data?.data ?? [],
		categories: categoriesRes.data?.data ?? [],
		items: itemsRes.data?.data ?? []
	};
};

export const actions: Actions = {
	checkout: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const branch_id = Number(form.get('branch_id'));
		const table_id_raw = form.get('table_id');
		const table_id = table_id_raw ? Number(table_id_raw) : undefined;
		const order_type = String(form.get('order_type') ?? 'dine_in') as
			| 'dine_in'
			| 'takeaway'
			| 'delivery';
		const notes = String(form.get('notes') ?? '').trim() || undefined;
		const itemsJson = String(form.get('items') ?? '[]');

		if (!branch_id) return fail(400, { error: 'Pilih cabang terlebih dahulu.' });

		let items: { menu_item_id: number; quantity: number; notes?: string }[] = [];
		try {
			items = JSON.parse(itemsJson);
		} catch {
			return fail(400, { error: 'Data item pesanan tidak valid.' });
		}

		if (items.length === 0) {
			return fail(400, { error: 'Tambahkan minimal 1 item ke pesanan.' });
		}

		const res = await api.orders.create(locals.accessToken, {
			branch_id,
			table_id,
			order_type,
			notes,
			items
		});

		if (res.error) return fail(400, { error: res.error });
		const orderId = res.data?.id;
		if (!orderId) return fail(400, { error: 'Gagal membuat pesanan.' });

		const invoiceRes = await api.payments.createInvoice(locals.accessToken, orderId);
		if (invoiceRes.error) {
			const paymentWarning = 'Invoice QRIS belum tersedia. Silakan coba buat invoice ulang dari modul pembayaran.';
			return {
				success: true,
				orderCode: res.data?.order_code,
				paymentWarning
			};
		}
		return { success: true, orderCode: res.data?.order_code, payment: invoiceRes.data };
	},

	loadTables: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });
		const form = await request.formData();
		const branch_id = Number(form.get('branch_id'));
		if (!branch_id) return fail(400, { error: 'branch_id wajib diisi' });

		const res = await api.tables.list(locals.accessToken, branch_id);
		if (res.error) return fail(400, { error: res.error });
		return { success: true, tables: res.data?.data ?? [] };
	}
};
