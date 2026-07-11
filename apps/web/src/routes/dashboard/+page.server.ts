import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { api } from '$lib/api/client';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.user || !locals.accessToken) {
		throw redirect(302, '/');
	}

	const [branchRes, categoriesRes, itemsRes] = await Promise.all([
		api.branches.list(locals.accessToken),
		api.menu.categories.list(locals.accessToken),
		api.menu.items.list(locals.accessToken)
	]);

	return {
		branches: branchRes.data?.data ?? [],
		categories: categoriesRes.data?.data ?? [],
		menuItems: itemsRes.data?.data ?? []
	};
};
