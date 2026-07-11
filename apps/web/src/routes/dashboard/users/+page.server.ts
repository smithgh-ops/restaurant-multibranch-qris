import type { PageServerLoad, Actions } from './$types';
import { redirect, fail } from '@sveltejs/kit';
import { api } from '$lib/api/client';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.user || !locals.accessToken) {
		throw redirect(302, '/');
	}

	const [usersRes, rolesRes, branchesRes] = await Promise.all([
		api.users.list(locals.accessToken),
		api.roles(locals.accessToken),
		api.branches.list(locals.accessToken)
	]);

	return {
		users: usersRes.data?.data ?? [],
		roles: rolesRes.data?.data ?? [],
		branches: branchesRes.data?.data ?? [],
		error: usersRes.error
	};
};

export const actions: Actions = {
	create: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const name = String(form.get('name') ?? '').trim();
		const email = String(form.get('email') ?? '').trim();
		const password = String(form.get('password') ?? '').trim();
		const roleId = Number(form.get('role_id'));
		const branchIdRaw = form.get('branch_id');
		const branchId = branchIdRaw ? Number(branchIdRaw) : undefined;

		if (!name || !email || !password) {
			return fail(400, { error: 'Nama, email, dan password wajib diisi.' });
		}

		const roles = roleId ? [{ role_id: roleId, branch_id: branchId }] : [];
		const res = await api.users.create(locals.accessToken, { name, email, password, roles });
		if (res.error) {
			return fail(400, { error: res.error });
		}
		return { success: true, action: 'create' };
	},

	update: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const id = Number(form.get('id'));
		const name = String(form.get('name') ?? '').trim() || undefined;
		const email = String(form.get('email') ?? '').trim() || undefined;
		const isActiveRaw = form.get('is_active');
		const is_active = isActiveRaw !== null ? isActiveRaw === 'true' : undefined;

		if (!id) return fail(400, { error: 'ID pengguna tidak valid.' });

		const res = await api.users.update(locals.accessToken, id, { name, email, is_active });
		if (res.error) {
			return fail(400, { error: res.error });
		}
		return { success: true, action: 'update' };
	},

	setRoles: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const id = Number(form.get('id'));
		const roleId = Number(form.get('role_id'));
		const branchIdRaw = form.get('branch_id');
		const branchId = branchIdRaw ? Number(branchIdRaw) : undefined;

		if (!id) return fail(400, { error: 'ID pengguna tidak valid.' });

		const roles = roleId ? [{ role_id: roleId, branch_id: branchId }] : [];
		const res = await api.users.setRoles(locals.accessToken, id, { roles });
		if (res.error) {
			return fail(400, { error: res.error });
		}
		return { success: true, action: 'setRoles' };
	},

	deactivate: async ({ locals, request }) => {
		if (!locals.accessToken) return fail(401, { error: 'Tidak terautentikasi' });

		const form = await request.formData();
		const id = Number(form.get('id'));
		if (!id) return fail(400, { error: 'ID pengguna tidak valid.' });

		const res = await api.users.deactivate(locals.accessToken, id);
		if (res.error) {
			return fail(400, { error: res.error });
		}
		return { success: true, action: 'deactivate' };
	}
};
