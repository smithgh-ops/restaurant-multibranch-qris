import type { PageServerLoad, Actions } from './$types';
import { redirect, fail } from '@sveltejs/kit';
import { api } from '$lib/api/client';

export const load: PageServerLoad = async ({ locals, url }) => {
	if (!locals.user || !locals.accessToken) {
		throw redirect(302, '/');
	}

	const branchesRes = await api.branches.list(locals.accessToken);
	const branches = branchesRes.data?.data ?? [];

	const selectedBranchId = Number(url.searchParams.get('branch_id')) || null;
	let tables: Awaited<ReturnType<typeof api.tables.list>>['data'] = undefined;

	if (selectedBranchId) {
		const tablesRes = await api.tables.list(locals.accessToken, selectedBranchId);
		tables = tablesRes.data;
	}

	return {
		branches,
		tables: tables?.data ?? [],
		selectedBranchId,
		error: branchesRes.error
	};
};

export const actions: Actions = {
	generateQR: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const branchId = Number(form.get('branch_id'));
		const tableId = Number(form.get('table_id'));

		if (!branchId || !tableId) return fail(400, { error: 'branch_id dan table_id wajib ada.' });

		const res = await api.selforder.generateToken(locals.accessToken, branchId, tableId);
		if (res.error) return fail(400, { error: res.error });
		return { success: true, token: res.data };
	}
};
