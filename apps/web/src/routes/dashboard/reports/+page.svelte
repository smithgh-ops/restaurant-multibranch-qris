<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	// ── Filter state ──────────────────────────────────────────────────────────
	const today = new Date().toISOString().slice(0, 10);
	const thirtyDaysAgo = new Date(Date.now() - 29 * 86400_000).toISOString().slice(0, 10);

	let selectedBranch = $state(String(data.selectedBranchId ?? ''));
	let dateFrom = $state(data.dateFrom ?? thirtyDaysAgo);
	let dateTo = $state(data.dateTo ?? today);

	function applyFilters() {
		const params = new URLSearchParams();
		if (selectedBranch) params.set('branch_id', selectedBranch);
		if (dateFrom) params.set('date_from', dateFrom);
		if (dateTo) params.set('date_to', dateTo);
		goto(`/dashboard/reports?${params.toString()}`, { replaceState: true });
	}

	// ── Helpers ───────────────────────────────────────────────────────────────
	function formatIDR(value: string | number): string {
		const n = typeof value === 'string' ? parseFloat(value) : value;
		return new Intl.NumberFormat('id-ID', {
			style: 'currency',
			currency: 'IDR',
			maximumFractionDigits: 0
		}).format(n);
	}

	const summary = $derived(data.sales?.summary);
	const dailyRows = $derived(data.sales?.daily ?? []);
	const topItems = $derived(data.topItems ?? []);

	// Build the export URL including the current auth token via the server-provided URL.
	// The CSV export endpoint requires a JWT so we use the server-side URL that already
	// has the correct API base baked in.
	const exportURL = $derived(data.exportURL);
</script>

<svelte:head>
	<title>Laporan & Analitik — RestoQRIS</title>
</svelte:head>

<div class="p-4 md:p-6 max-w-6xl mx-auto space-y-6">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Laporan & Analitik</h1>
			<p class="text-sm text-gray-500 mt-0.5">Ringkasan penjualan dan performa cabang</p>
		</div>
		<a
			href={exportURL}
			class="inline-flex items-center gap-2 px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 transition-colors"
			download
		>
			⬇️ Ekspor CSV
		</a>
	</div>

	<!-- Filters -->
	<div class="bg-white rounded-xl shadow-sm border border-gray-100 p-4">
		<div class="grid grid-cols-1 sm:grid-cols-4 gap-3 items-end">
			<!-- Branch selector -->
			<div>
				<label for="filter-branch" class="block text-xs font-medium text-gray-600 mb-1">Cabang</label>
				<select
					id="filter-branch"
					bind:value={selectedBranch}
					class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
				>
					<option value="">Semua Cabang</option>
					{#each data.branches as branch}
						<option value={String(branch.id)}>{branch.name}</option>
					{/each}
				</select>
			</div>
			<!-- Date from -->
			<div>
				<label for="filter-date-from" class="block text-xs font-medium text-gray-600 mb-1">Dari Tanggal</label>
				<input
					id="filter-date-from"
					type="date"
					bind:value={dateFrom}
					class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
				/>
			</div>
			<!-- Date to -->
			<div>
				<label for="filter-date-to" class="block text-xs font-medium text-gray-600 mb-1">Sampai Tanggal</label>
				<input
					id="filter-date-to"
					type="date"
					bind:value={dateTo}
					class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
				/>
			</div>
			<!-- Apply button -->
			<button
				onclick={applyFilters}
				class="px-4 py-2 bg-orange-500 text-white text-sm font-medium rounded-lg hover:bg-orange-600 transition-colors"
			>
				Terapkan
			</button>
		</div>
	</div>

	<!-- Summary cards -->
	{#if summary}
		<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
			<div class="bg-white rounded-xl shadow-sm border border-gray-100 p-5">
				<p class="text-sm text-gray-500">Total Pendapatan</p>
				<p class="text-2xl font-bold text-gray-900 mt-1">{formatIDR(summary.total_revenue)}</p>
			</div>
			<div class="bg-white rounded-xl shadow-sm border border-gray-100 p-5">
				<p class="text-sm text-gray-500">Jumlah Pesanan</p>
				<p class="text-2xl font-bold text-gray-900 mt-1">{summary.order_count.toLocaleString('id-ID')}</p>
			</div>
			<div class="bg-white rounded-xl shadow-sm border border-gray-100 p-5">
				<p class="text-sm text-gray-500">Rata-Rata Nilai Pesanan</p>
				<p class="text-2xl font-bold text-gray-900 mt-1">{formatIDR(summary.avg_order_value)}</p>
			</div>
		</div>
	{/if}

	<!-- Daily sales table -->
	<div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
		<div class="px-5 py-4 border-b border-gray-100">
			<h2 class="text-base font-semibold text-gray-800">Penjualan Harian</h2>
		</div>
		{#if dailyRows.length === 0}
			<p class="text-sm text-gray-400 px-5 py-8 text-center">Tidak ada data untuk periode ini.</p>
		{:else}
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-100 text-sm">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-5 py-3 text-left font-medium text-gray-500 uppercase tracking-wide text-xs">Tanggal</th>
							<th class="px-5 py-3 text-right font-medium text-gray-500 uppercase tracking-wide text-xs">Pesanan</th>
							<th class="px-5 py-3 text-right font-medium text-gray-500 uppercase tracking-wide text-xs">Pendapatan</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-50">
						{#each dailyRows as row}
							<tr class="hover:bg-gray-50 transition-colors">
								<td class="px-5 py-3 text-gray-700">{row.date}</td>
								<td class="px-5 py-3 text-right text-gray-700">{row.order_count}</td>
								<td class="px-5 py-3 text-right font-medium text-gray-900">{formatIDR(row.revenue)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>

	<!-- Top items table -->
	<div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
		<div class="px-5 py-4 border-b border-gray-100">
			<h2 class="text-base font-semibold text-gray-800">Produk Terlaris (Top 10)</h2>
		</div>
		{#if topItems.length === 0}
			<p class="text-sm text-gray-400 px-5 py-8 text-center">Tidak ada data untuk periode ini.</p>
		{:else}
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-100 text-sm">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-5 py-3 text-left font-medium text-gray-500 uppercase tracking-wide text-xs">#</th>
							<th class="px-5 py-3 text-left font-medium text-gray-500 uppercase tracking-wide text-xs">Item</th>
							<th class="px-5 py-3 text-right font-medium text-gray-500 uppercase tracking-wide text-xs">Qty Terjual</th>
							<th class="px-5 py-3 text-right font-medium text-gray-500 uppercase tracking-wide text-xs">Total Pendapatan</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-50">
						{#each topItems as item, i}
							<tr class="hover:bg-gray-50 transition-colors">
								<td class="px-5 py-3 text-gray-400 font-medium">{i + 1}</td>
								<td class="px-5 py-3 text-gray-800 font-medium">{item.item_name}</td>
								<td class="px-5 py-3 text-right text-gray-700">{item.total_qty.toLocaleString('id-ID')}</td>
								<td class="px-5 py-3 text-right font-medium text-gray-900">{formatIDR(item.total_revenue)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>
