<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData, ActionData } from './$types';
	import type { UserWithRoles } from '$lib/api/client';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let showCreateForm = $state(false);
	let editingUser = $state<UserWithRoles | null>(null);
	let editingRolesUser = $state<UserWithRoles | null>(null);
</script>

<svelte:head>
	<title>Manajemen Pengguna — RestoQRIS</title>
</svelte:head>

<div class="max-w-5xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h2 class="text-2xl font-bold text-gray-800">Manajemen Pengguna</h2>
			<p class="mt-1 text-gray-500 text-sm">Kelola staf dan hak akses mereka.</p>
		</div>
		<button
			onclick={() => {
				showCreateForm = !showCreateForm;
				editingUser = null;
				editingRolesUser = null;
			}}
			class="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg transition-colors"
		>
			{showCreateForm ? 'Batal' : '+ Tambah Pengguna'}
		</button>
	</div>

	<!-- Feedback messages -->
	{#if form && !form.success && form.error}
		<div class="mb-4 rounded-lg bg-red-50 border border-red-200 text-red-700 px-4 py-3 text-sm">
			{form.error}
		</div>
	{/if}
	{#if form && form.success}
		<div class="mb-4 rounded-lg bg-green-50 border border-green-200 text-green-700 px-4 py-3 text-sm">
			Berhasil disimpan.
		</div>
	{/if}

	<!-- Create Form -->
	{#if showCreateForm}
		<div class="bg-white rounded-xl shadow-sm p-6 mb-6">
			<h3 class="text-lg font-semibold text-gray-800 mb-4">Tambah Pengguna Baru</h3>
			<form
				method="POST"
				action="?/create"
				use:enhance={() => {
					return async ({ update }) => {
						await update();
						if (form?.success) showCreateForm = false;
					};
				}}
				class="space-y-4"
			>
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="new-name">Nama</label>
						<input
							id="new-name"
							name="name"
							type="text"
							required
							placeholder="Nama Lengkap"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="new-email">Email</label>
						<input
							id="new-email"
							name="email"
							type="email"
							required
							placeholder="email@contoh.com"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="new-password">Password Sementara</label>
						<input
							id="new-password"
							name="password"
							type="password"
							required
							minlength="6"
							placeholder="Min. 6 karakter"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="new-role">Role</label>
						<select
							id="new-role"
							name="role_id"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						>
							<option value="">— Pilih Role —</option>
							{#each data.roles as role}
								<option value={role.id}>{role.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="new-branch">Cabang (opsional)</label>
						<select
							id="new-branch"
							name="branch_id"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						>
							<option value="">— Global (semua cabang) —</option>
							{#each data.branches as branch}
								<option value={branch.id}>{branch.name}</option>
							{/each}
						</select>
					</div>
				</div>
				<div class="flex gap-3 pt-2">
					<button
						type="submit"
						class="px-5 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg transition-colors"
					>
						Simpan
					</button>
					<button
						type="button"
						onclick={() => (showCreateForm = false)}
						class="px-5 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 text-sm font-semibold rounded-lg transition-colors"
					>
						Batal
					</button>
				</div>
			</form>
		</div>
	{/if}

	<!-- Edit User Form -->
	{#if editingUser}
		<div class="bg-white rounded-xl shadow-sm p-6 mb-6">
			<h3 class="text-lg font-semibold text-gray-800 mb-4">Edit Pengguna: {editingUser.name}</h3>
			<form
				method="POST"
				action="?/update"
				use:enhance={() => {
					return async ({ update }) => {
						await update();
						if (form?.success) editingUser = null;
					};
				}}
				class="space-y-4"
			>
				<input type="hidden" name="id" value={editingUser.id} />
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="edit-name">Nama</label>
						<input
							id="edit-name"
							name="name"
							type="text"
							value={editingUser.name}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="edit-email">Email</label>
						<input
							id="edit-email"
							name="email"
							type="email"
							value={editingUser.email}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="edit-active">Status</label>
						<select
							id="edit-active"
							name="is_active"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						>
							<option value="true" selected={editingUser.is_active}>Aktif</option>
							<option value="false" selected={!editingUser.is_active}>Nonaktif</option>
						</select>
					</div>
				</div>
				<div class="flex gap-3 pt-2">
					<button
						type="submit"
						class="px-5 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg transition-colors"
					>
						Simpan
					</button>
					<button
						type="button"
						onclick={() => (editingUser = null)}
						class="px-5 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 text-sm font-semibold rounded-lg transition-colors"
					>
						Batal
					</button>
				</div>
			</form>
		</div>
	{/if}

	<!-- Edit Roles Form -->
	{#if editingRolesUser}
		<div class="bg-white rounded-xl shadow-sm p-6 mb-6">
			<h3 class="text-lg font-semibold text-gray-800 mb-4">Ubah Role: {editingRolesUser.name}</h3>
			<form
				method="POST"
				action="?/setRoles"
				use:enhance={() => {
					return async ({ update }) => {
						await update();
						if (form?.success) editingRolesUser = null;
					};
				}}
				class="space-y-4"
			>
				<input type="hidden" name="id" value={editingRolesUser.id} />
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="role-id">Role</label>
						<select
							id="role-id"
							name="role_id"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						>
							<option value="">— Hapus Semua Role —</option>
							{#each data.roles as role}
								<option
									value={role.id}
									selected={editingRolesUser.roles.some((r) => r.role_id === role.id)}
								>{role.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="role-branch">Cabang</label>
						<select
							id="role-branch"
							name="branch_id"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						>
							<option value="">— Global (semua cabang) —</option>
							{#each data.branches as branch}
								<option
									value={branch.id}
									selected={editingRolesUser.roles.some((r) => r.branch_id === branch.id)}
								>{branch.name}</option>
							{/each}
						</select>
					</div>
				</div>
				<div class="flex gap-3 pt-2">
					<button
						type="submit"
						class="px-5 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg transition-colors"
					>
						Simpan
					</button>
					<button
						type="button"
						onclick={() => (editingRolesUser = null)}
						class="px-5 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 text-sm font-semibold rounded-lg transition-colors"
					>
						Batal
					</button>
				</div>
			</form>
		</div>
	{/if}

	<!-- Users Table -->
	<div class="bg-white rounded-xl shadow-sm overflow-hidden">
		{#if data.users.length === 0}
			<div class="px-6 py-12 text-center text-gray-400 text-sm">Belum ada pengguna.</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-100">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Nama</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Email</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Role & Cabang</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
							<th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Aksi</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-50">
						{#each data.users as u}
							<tr class="hover:bg-gray-50 transition-colors">
								<td class="px-4 py-3 text-sm font-medium text-gray-800">{u.name}</td>
								<td class="px-4 py-3 text-sm text-gray-600">{u.email}</td>
								<td class="px-4 py-3 text-sm text-gray-600">
									{#if u.roles.length > 0}
										<div class="flex flex-wrap gap-1">
											{#each u.roles as r}
												<span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs bg-orange-50 text-orange-700 border border-orange-100">
													{r.role_name}
													{#if r.branch_id}
														{@const b = data.branches.find((br) => br.id === r.branch_id)}
														{#if b}<span class="text-gray-400">/ {b.name}</span>{/if}
													{:else}
														<span class="text-gray-400">/ global</span>
													{/if}
												</span>
											{/each}
										</div>
									{:else}
										<span class="text-gray-400 text-xs italic">Tidak ada role</span>
									{/if}
								</td>
								<td class="px-4 py-3 text-sm">
									{#if u.is_active}
										<span class="inline-flex px-2 py-0.5 rounded-full text-xs font-medium bg-green-50 text-green-700">Aktif</span>
									{:else}
										<span class="inline-flex px-2 py-0.5 rounded-full text-xs font-medium bg-red-50 text-red-600">Nonaktif</span>
									{/if}
								</td>
								<td class="px-4 py-3 text-right text-sm">
									<div class="flex items-center justify-end gap-2">
										<button
											onclick={() => {
												editingUser = u;
												editingRolesUser = null;
												showCreateForm = false;
											}}
											class="text-orange-600 hover:text-orange-800 text-xs font-medium"
										>Edit</button>
										<button
											onclick={() => {
												editingRolesUser = u;
												editingUser = null;
												showCreateForm = false;
											}}
											class="text-blue-600 hover:text-blue-800 text-xs font-medium"
										>Role</button>
										{#if u.is_active}
											<form method="POST" action="?/deactivate" use:enhance>
												<input type="hidden" name="id" value={u.id} />
												<button
													type="submit"
													class="text-red-500 hover:text-red-700 text-xs font-medium"
													onclick={(e) => {
														if (!confirm(`Nonaktifkan ${u.name}?`)) e.preventDefault();
													}}
												>Nonaktifkan</button>
											</form>
										{/if}
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>
