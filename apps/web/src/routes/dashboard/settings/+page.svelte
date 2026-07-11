<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData, ActionData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<svelte:head>
	<title>Pengaturan — RestoQRIS</title>
</svelte:head>

<div class="max-w-2xl mx-auto space-y-6">
	<div class="mb-2">
		<h2 class="text-2xl font-bold text-gray-800">Pengaturan</h2>
		<p class="mt-1 text-gray-500 text-sm">Profil organisasi dan keamanan akun.</p>
	</div>

	<!-- Session info -->
	{#if data.me}
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h3 class="text-base font-semibold text-gray-800 mb-4">Sesi Aktif</h3>
			<dl class="divide-y divide-gray-100">
				<div class="py-2 flex justify-between text-sm">
					<dt class="text-gray-500">Nama</dt>
					<dd class="text-gray-800 font-medium">{data.me.name}</dd>
				</div>
				<div class="py-2 flex justify-between text-sm">
					<dt class="text-gray-500">Email</dt>
					<dd class="text-gray-800">{data.me.email}</dd>
				</div>
				<div class="py-2 flex justify-between text-sm">
					<dt class="text-gray-500">Role</dt>
					<dd class="text-gray-800">
						{#if data.me.roles.length > 0}
							<div class="flex flex-wrap gap-1 justify-end">
								{#each data.me.roles as r}
									<span class="inline-flex px-2 py-0.5 rounded-full text-xs bg-orange-50 text-orange-700 border border-orange-100">
										{r.role_name}
									</span>
								{/each}
							</div>
						{:else}
							<span class="text-gray-400 italic">Tidak ada role</span>
						{/if}
					</dd>
				</div>
			</dl>
		</div>
	{/if}

	<!-- Organization settings -->
	<div class="bg-white rounded-xl shadow-sm p-6">
		<h3 class="text-base font-semibold text-gray-800 mb-4">Profil Organisasi</h3>

		{#if form && form.section === 'org'}
			{#if form.success}
				<div class="mb-4 rounded-lg bg-green-50 border border-green-200 text-green-700 px-4 py-3 text-sm">
					Profil organisasi berhasil diperbarui.
				</div>
			{:else if form.error}
				<div class="mb-4 rounded-lg bg-red-50 border border-red-200 text-red-700 px-4 py-3 text-sm">
					{form.error}
				</div>
			{/if}
		{/if}

		<form method="POST" action="?/updateOrg" use:enhance class="space-y-4">
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1" for="org-name">Nama Organisasi</label>
				<input
					id="org-name"
					name="name"
					type="text"
					value={data.organization?.name ?? ''}
					placeholder="Nama Restoran"
					class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1" for="org-slug">Slug</label>
				<input
					id="org-slug"
					name="slug"
					type="text"
					value={data.organization?.slug ?? ''}
					placeholder="nama-restoran"
					class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
				/>
				<p class="mt-1 text-xs text-gray-400">Slug digunakan sebagai pengenal unik URL-friendly.</p>
			</div>
			<button
				type="submit"
				class="px-5 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg transition-colors"
			>
				Simpan Profil
			</button>
		</form>
	</div>

	<!-- Change password -->
	<div class="bg-white rounded-xl shadow-sm p-6">
		<h3 class="text-base font-semibold text-gray-800 mb-4">Ganti Password</h3>

		{#if form && form.section === 'password'}
			{#if form.success}
				<div class="mb-4 rounded-lg bg-green-50 border border-green-200 text-green-700 px-4 py-3 text-sm">
					Password berhasil diubah.
				</div>
			{:else if form.error}
				<div class="mb-4 rounded-lg bg-red-50 border border-red-200 text-red-700 px-4 py-3 text-sm">
					{form.error}
				</div>
			{/if}
		{/if}

		<form method="POST" action="?/changePassword" use:enhance class="space-y-4">
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1" for="current-pass">Password Saat Ini</label>
				<input
					id="current-pass"
					name="current_password"
					type="password"
					required
					placeholder="••••••••"
					class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1" for="new-pass">Password Baru</label>
				<input
					id="new-pass"
					name="new_password"
					type="password"
					required
					minlength="6"
					placeholder="Min. 6 karakter"
					class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1" for="confirm-pass">Konfirmasi Password Baru</label>
				<input
					id="confirm-pass"
					name="confirm_password"
					type="password"
					required
					minlength="6"
					placeholder="Ulangi password baru"
					class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
				/>
			</div>
			<button
				type="submit"
				class="px-5 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg transition-colors"
			>
				Ganti Password
			</button>
		</form>
	</div>
</div>
