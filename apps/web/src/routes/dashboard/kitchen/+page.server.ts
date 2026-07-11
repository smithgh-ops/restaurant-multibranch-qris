import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	const destination = new URL('/dashboard/kds', url);
	if (url.searchParams.get('branch_id')) {
		destination.searchParams.set('branch_id', url.searchParams.get('branch_id') ?? '');
	}
	if (url.searchParams.get('station_id')) {
		destination.searchParams.set('station_id', url.searchParams.get('station_id') ?? '');
	}
	throw redirect(302, `${destination.pathname}${destination.search}`);
};
