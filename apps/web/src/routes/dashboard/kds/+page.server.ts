import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { api } from '$lib/api/client';
import type { KitchenTicketStatus } from '$lib/api/client';

export const load: PageServerLoad = async ({ locals, url }) => {
	if (!locals.user || !locals.accessToken) {
		throw redirect(302, '/');
	}

	const branchesRes = await api.branches.list(locals.accessToken);
	const branches = branchesRes.data?.data ?? [];
	const requestedBranchId = url.searchParams.get('branch_id')
		? Number(url.searchParams.get('branch_id'))
		: undefined;
	const selectedBranchId = requestedBranchId ?? branches[0]?.id ?? null;
	const selectedStationId = url.searchParams.get('station_id')
		? Number(url.searchParams.get('station_id'))
		: null;

	if (!selectedBranchId) {
		return {
			branches,
			stations: [],
			tickets: [],
			selectedBranchId: null,
			selectedStationId,
			wsToken: null
		};
	}

	const [stationsRes, ticketsRes] = await Promise.all([
		api.kds.stations.list(locals.accessToken, selectedBranchId),
		api.kds.tickets.list(locals.accessToken, selectedBranchId, {
			station_id: selectedStationId ?? undefined
		})
	]);

	return {
		branches,
		stations: stationsRes.data?.data ?? [],
		tickets: ticketsRes.data?.data ?? [],
		selectedBranchId,
		selectedStationId,
		wsToken: locals.accessToken
	};
};

export const actions: Actions = {
	createStation: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const branchId = Number(form.get('branch_id'));
		const name = String(form.get('name') ?? '').trim();
		if (!branchId || !name) return fail(400, { error: 'Cabang dan nama station wajib diisi.' });

		const res = await api.kds.stations.create(locals.accessToken, branchId, { name });
		if (res.error) return fail(400, { error: res.error });
		return { success: true };
	},

	updateTicketStatus: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const branchId = Number(form.get('branch_id'));
		const ticketId = Number(form.get('ticket_id'));
		const status = String(form.get('status')) as KitchenTicketStatus;
		if (!branchId || !ticketId || !status) {
			return fail(400, { error: 'Data perubahan status tidak valid.' });
		}

		const res = await api.kds.tickets.updateStatus(locals.accessToken, branchId, ticketId, status);
		if (res.error) return fail(400, { error: res.error });
		return { success: true };
	}
};
