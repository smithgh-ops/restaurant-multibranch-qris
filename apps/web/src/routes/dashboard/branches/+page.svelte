<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import type { PageData, ActionData } from './$types';
	import type { Branch } from '$lib/api/client';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let showCreateForm = $state(false);
	let editingBranch = $state<Branch | null>(null);
	let selectedActiveFilter = $state(data.activeFilter);

	function slugify(name: string) {
		return name
			.toLowerCase()
			.replace(/\s+/g, '-')
			.replace(/[^a-z0-9-]/g, '');
	}

	let newName = $state('');
	let newSlug = $state('');
</script>

<svelte:head>
	<title>Manajemen Cabang — RestoQRIS</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h2 class="text-2xl font-bold text-gray-800">Manajemen Cabang</h2>
			<p class="mt-1 text-gray-500 text-sm">Kelola cabang restoran Anda.</p>
		</div>
		<div class="flex items-center gap-3">
			<select
				bind:value={selectedActiveFilter}
				onchange={() => {
					const params = new URLSearchParams(window.location.search);
					if (selectedActiveFilter === 'all') {
						params.delete('active');
					} else {
						params.set('active', selectedActiveFilter);
					}
					const query = params.toString();
					goto(query ? `?${query}` : '?', { invalidateAll: true });
				}}
				class="rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
			>
				<option value="all">Semua Status</option>
				<option value="true">Aktif</option>
				<option value="false">Nonaktif</option>
			</select>
			<button
				onclick={() => {
					showCreateForm = !showCreateForm;
					editingBranch = null;
				}}
				class="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg transition-colors"
			>
				{showCreateForm ? 'Batal' : '+ Tambah Cabang'}
			</button>
		</div>
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

	<!-- Create form -->
	{#if showCreateForm}
		<div class="bg-white rounded-xl shadow-sm p-6 mb-6">
			<h3 class="text-lg font-semibold text-gray-800 mb-4">Tambah Cabang Baru</h3>
			<form
				method="POST"
				action="?/create"
				use:enhance={() => {
					return async ({ update }) => {
						await update();
						if (form?.success) {
							showCreateForm = false;
							newName = '';
							newSlug = '';
						}
					};
				}}
				class="space-y-4"
			>
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="new-name">Nama Cabang</label>
						<input
							id="new-name"
							name="name"
							type="text"
							required
							bind:value={newName}
							oninput={() => { newSlug = slugify(newName); }}
							placeholder="Cabang Utama"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="new-slug">Slug</label>
						<input
							id="new-slug"
							name="slug"
							type="text"
							required
							bind:value={newSlug}
							placeholder="cabang-utama"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="new-address">Alamat</label>
						<input
							id="new-address"
							name="address"
							type="text"
							placeholder="Jl. Contoh No. 1"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="new-phone">Telepon</label>
						<input
							id="new-phone"
							name="phone"
							type="text"
							placeholder="08xxxxxxxxxx"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
				</div>
				<div class="flex gap-3">
					<button
						type="submit"
						class="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg"
					>
						Simpan
					</button>
					<button
						type="button"
						onclick={() => (showCreateForm = false)}
						class="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 text-sm font-semibold rounded-lg"
					>
						Batal
					</button>
				</div>
			</form>
		</div>
	{/if}

	<!-- Edit form -->
	{#if editingBranch}
		<div class="bg-white rounded-xl shadow-sm p-6 mb-6">
			<h3 class="text-lg font-semibold text-gray-800 mb-4">Edit Cabang: {editingBranch.name}</h3>
			<form
				method="POST"
				action="?/update"
				use:enhance={() => {
					return async ({ update }) => {
						await update();
						if (form?.success) editingBranch = null;
					};
				}}
				class="space-y-4"
			>
				<input type="hidden" name="id" value={editingBranch.id} />
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="edit-name">Nama Cabang</label>
						<input
							id="edit-name"
							name="name"
							type="text"
							value={editingBranch.name}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="edit-slug">Slug</label>
						<input
							id="edit-slug"
							name="slug"
							type="text"
							value={editingBranch.slug}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="edit-address">Alamat</label>
						<input
							id="edit-address"
							name="address"
							type="text"
							value={editingBranch.address ?? ''}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1" for="edit-phone">Telepon</label>
						<input
							id="edit-phone"
							name="phone"
							type="text"
							value={editingBranch.phone ?? ''}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
					</div>
				</div>
				<div>
					<label for="edit-branch-status" class="block text-sm font-medium text-gray-700 mb-1">Status</label>
					<select
						id="edit-branch-status"
						name="is_active"
						class="rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
					>
						<option value="true" selected={editingBranch.is_active}>Aktif</option>
						<option value="false" selected={!editingBranch.is_active}>Nonaktif</option>
					</select>
				</div>
				<div class="flex gap-3">
					<button
						type="submit"
						class="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg"
					>
						Simpan Perubahan
					</button>
					<button
						type="button"
						onclick={() => (editingBranch = null)}
						class="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 text-sm font-semibold rounded-lg"
					>
						Batal
					</button>
				</div>
			</form>
		</div>
	{/if}

	<!-- Branch list -->
	{#if data.error}
		<div class="bg-red-50 border border-red-200 rounded-xl p-6 text-red-700 text-sm">
			Gagal memuat data: {data.error}
		</div>
	{:else if data.branches.length === 0}
		<div class="bg-white rounded-xl shadow-sm p-12 text-center">
			<span class="text-5xl mb-4 block">🏪</span>
			<p class="text-gray-500">Belum ada cabang. Tambahkan cabang pertama Anda.</p>
		</div>
	{:else}
		<div class="space-y-3">
			{#each data.branches as branch}
				<div class="bg-white rounded-xl shadow-sm p-5 flex items-center justify-between gap-4">
					<div class="flex items-center gap-4">
						<span class="text-2xl">🏪</span>
						<div>
							<p class="font-semibold text-gray-800">{branch.name}</p>
							<p class="text-xs text-gray-400">/{branch.slug}</p>
							{#if branch.address}
								<p class="text-xs text-gray-500 mt-0.5">{branch.address}</p>
							{/if}
							{#if branch.phone}
								<p class="text-xs text-gray-500">{branch.phone}</p>
							{/if}
						</div>
					</div>
					<div class="flex items-center gap-3 shrink-0">
						<span
							class="text-xs font-medium px-2 py-0.5 rounded-full {branch.is_active
								? 'bg-green-50 text-green-600'
								: 'bg-gray-100 text-gray-400'}"
						>
							{branch.is_active ? 'Aktif' : 'Nonaktif'}
						</span>
						<button
							onclick={() => {
								editingBranch = branch;
								showCreateForm = false;
							}}
							class="text-sm text-orange-500 hover:text-orange-700 font-medium"
						>
							Edit
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
