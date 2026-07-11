import type { PageServerLoad, Actions } from './$types';
import { redirect, fail } from '@sveltejs/kit';
import { api } from '$lib/api/client';
import type { OrderStatus } from '$lib/api/client';

export const load: PageServerLoad = async ({ locals, url }) => {
	if (!locals.user || !locals.accessToken) {
		throw redirect(302, '/');
	}

	const branchId = url.searchParams.get('branch_id')
		? Number(url.searchParams.get('branch_id'))
		: undefined;
	const status = (url.searchParams.get('status') as OrderStatus | null) ?? undefined;

	const [branchesRes, ordersRes] = await Promise.all([
		api.branches.list(locals.accessToken),
		api.orders.list(locals.accessToken, {
			branch_id: branchId,
			status
		})
	]);

	return {
		branches: branchesRes.data?.data ?? [],
		orders: ordersRes.data?.data ?? [],
		selectedBranchId: branchId ?? null,
		selectedStatus: status ?? null
	};
};

export const actions: Actions = {
	updateStatus: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const id = Number(form.get('order_id'));
		const status = String(form.get('status')) as OrderStatus;

		if (!id || !status) return fail(400, { error: 'Data tidak valid.' });

		const res = await api.orders.updateStatus(locals.accessToken, id, status);
		if (res.error) return fail(400, { error: res.error });
		return { success: true };
	}
};
