import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { api } from '$lib/api/client';

export const load: PageServerLoad = async ({ locals, url }) => {
	if (!locals.user || !locals.accessToken) {
		throw redirect(302, '/');
	}

	const token = locals.accessToken;

	const branchId = url.searchParams.get('branch_id')
		? Number(url.searchParams.get('branch_id'))
		: undefined;
	const dateFrom = url.searchParams.get('date_from') ?? undefined;
	const dateTo = url.searchParams.get('date_to') ?? undefined;

	const filters = { branch_id: branchId, date_from: dateFrom, date_to: dateTo };

	const [branchesRes, salesRes, topItemsRes] = await Promise.all([
		api.branches.list(token),
		api.reports.sales(token, filters),
		api.reports.topItems(token, { ...filters, limit: 10 })
	]);

	// Build export URL server-side so we can pass the base URL to the client.
	const exportURL = api.reports.exportURL(filters);

	return {
		branches: branchesRes.data?.data ?? [],
		sales: salesRes.data ?? null,
		topItems: topItemsRes.data?.data ?? [],
		selectedBranchId: branchId ?? null,
		dateFrom: dateFrom ?? null,
		dateTo: dateTo ?? null,
		exportURL
	};
};
