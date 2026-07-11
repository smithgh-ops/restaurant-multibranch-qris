<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import type { PageData, ActionData } from './$types';
	import type { QRTableToken, RestaurantTable } from '$lib/api/client';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let selectedBranchId = $state<number | null>(data.selectedBranchId);
	// Map of table_id → QR token (from generateQR action results)
	let tokenMap = $state<Record<number, QRTableToken>>({});

	const APP_URL = import.meta.env.VITE_APP_URL ?? 'http://localhost:3000';

	function selfOrderURL(token: string) {
		return `${APP_URL}/order/${token}`;
	}

	function qrImageURL(token: string) {
		// NOTE: qrserver.com is used for convenience in development/staging.
		// In production, consider a self-hosted or client-side QR library to avoid
		// sending the self-order URL to a third-party service.
		return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(selfOrderURL(token))}`;
	}

	$effect(() => {
		if (form && 'token' in form && form.token) {
			const t = form.token as QRTableToken;
			tokenMap[t.table_id] = t;
		}
	});

	function onBranchChange() {
		if (selectedBranchId) {
			goto(`/dashboard/tables?branch_id=${selectedBranchId}`, { replaceState: true, invalidateAll: true });
		}
	}

	function copyToClipboard(text: string) {
		navigator.clipboard.writeText(text).catch(() => {});
	}

	const tables = $derived(data.tables as RestaurantTable[]);
</script>

<svelte:head>
	<title>Manajemen Meja — RestoQRIS</title>
</svelte:head>

<div class="max-w-5xl mx-auto">
	<div class="mb-6">
		<h2 class="text-2xl font-bold text-gray-800">Manajemen Meja & QR Token</h2>
		<p class="mt-1 text-gray-500 text-sm">
			Generate QR code per meja untuk fitur self-order pelanggan.
		</p>
	</div>

	{#if form && !form.success && 'error' in form && form.error}
		<div class="mb-4 rounded-lg bg-red-50 border border-red-200 text-red-700 px-4 py-3 text-sm">
			{form.error}
		</div>
	{/if}

	<!-- Branch selector -->
	<div class="bg-white rounded-xl shadow-sm p-4 mb-6 flex items-end gap-4 flex-wrap">
		<div class="flex-1 min-w-[200px]">
			<label for="branch-select" class="block text-xs font-medium text-gray-500 mb-1">Pilih Cabang</label>
			<select
				id="branch-select"
				bind:value={selectedBranchId}
				onchange={onBranchChange}
				class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
			>
				<option value={null}>-- Pilih cabang --</option>
				{#each data.branches.filter((b) => b.is_active) as branch}
					<option value={branch.id}>{branch.name}</option>
				{/each}
			</select>
		</div>
	</div>

	<!-- Tables list -->
	{#if !selectedBranchId}
		<div class="bg-white rounded-xl shadow-sm p-12 text-center text-gray-400">
			<span class="text-4xl block mb-3">🪑</span>
			Pilih cabang untuk melihat daftar meja.
		</div>
	{:else if tables.length === 0}
		<div class="bg-white rounded-xl shadow-sm p-12 text-center text-gray-400">
			<span class="text-4xl block mb-3">🪑</span>
			Belum ada meja pada cabang ini. Tambahkan meja melalui halaman POS.
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each tables as table}
				{@const activeToken = tokenMap[table.id]}
				<div class="bg-white rounded-xl shadow-sm p-5 flex flex-col gap-4">
					<!-- Table header -->
					<div class="flex items-start justify-between">
						<div>
							<p class="font-semibold text-gray-800 text-lg">Meja {table.table_number}</p>
							<p class="text-xs text-gray-500">Kapasitas: {table.capacity} orang</p>
						</div>
						<span
							class="text-xs px-2 py-0.5 rounded-full {table.is_active
								? 'bg-green-50 text-green-600'
								: 'bg-gray-100 text-gray-400'}"
						>
							{table.is_active ? 'Aktif' : 'Nonaktif'}
						</span>
					</div>

					<!-- QR Token section -->
					{#if activeToken}
						<div class="flex flex-col items-center gap-2 border-t border-gray-100 pt-4">
							<img
								src={qrImageURL(activeToken.token)}
								alt={`QR code self-order untuk Meja ${table.table_number}`}
								class="w-40 h-40 rounded-lg border border-gray-200 bg-white p-1"
							/>
							<p class="text-xs text-gray-500 text-center break-all max-w-xs">
								{selfOrderURL(activeToken.token)}
							</p>
							<div class="flex gap-2 flex-wrap justify-center">
								<button
									onclick={() => copyToClipboard(selfOrderURL(activeToken.token))}
									class="px-3 py-1.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg font-medium"
								>
									📋 Salin URL
								</button>
								<a
									href={selfOrderURL(activeToken.token)}
									target="_blank"
									rel="noopener noreferrer"
									class="px-3 py-1.5 text-xs bg-orange-50 hover:bg-orange-100 text-orange-700 rounded-lg font-medium"
								>
									🔗 Buka
								</a>
							</div>
						</div>
					{:else}
						<div class="border-t border-gray-100 pt-4 text-center">
							<p class="text-sm text-gray-400 mb-3">Belum ada QR token aktif.</p>
						</div>
					{/if}

					<!-- Generate / regenerate button -->
					<form
						method="POST"
						action="?/generateQR"
						use:enhance={() => {
							return async ({ update }) => {
								await update({ reset: false });
							};
						}}
					>
						<input type="hidden" name="branch_id" value={selectedBranchId} />
						<input type="hidden" name="table_id" value={table.id} />
						<button
							type="submit"
							class="w-full py-2 px-4 text-sm font-semibold rounded-lg transition-colors
								{activeToken
								? 'bg-gray-100 hover:bg-gray-200 text-gray-700'
								: 'bg-orange-500 hover:bg-orange-600 text-white'}"
						>
							{activeToken ? '🔄 Generate Ulang QR' : '✨ Generate QR Token'}
						</button>
					</form>
				</div>
			{/each}
		</div>
	{/if}
</div>
