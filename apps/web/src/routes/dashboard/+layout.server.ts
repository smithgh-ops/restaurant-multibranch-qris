import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { api } from '$lib/api/client';

export const load: LayoutServerLoad = async ({ locals }) => {
	if (!locals.user || !locals.accessToken) {
		throw redirect(302, '/');
	}

	// Load organization and branch summary in parallel
	const [orgRes, branchRes] = await Promise.all([
		api.organization(locals.accessToken),
		api.branches.list(locals.accessToken)
	]);

	return {
		user: locals.user,
		organization: orgRes.data ?? null,
		branchCount: branchRes.data?.data?.length ?? 0
	};
};
