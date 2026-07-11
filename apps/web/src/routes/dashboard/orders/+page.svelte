<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import type { PageData, ActionData } from './$types';
	import type { Order, OrderStatus } from '$lib/api/client';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let filterBranch = $state(data.selectedBranchId);
	let filterStatus = $state<OrderStatus | ''>(data.selectedStatus ?? '');
	let expandedOrderId = $state<number | null>(null);

	const priceFormatter = new Intl.NumberFormat('id-ID', {
		style: 'currency',
		currency: 'IDR',
		minimumFractionDigits: 0
	});

	function formatPrice(p: string) {
		return priceFormatter.format(Number(p));
	}

	function applyFilters() {
		const params = new URLSearchParams();
		if (filterBranch) params.set('branch_id', String(filterBranch));
		if (filterStatus) params.set('status', filterStatus);
		goto(`/dashboard/orders${params.toString() ? '?' + params.toString() : ''}`);
	}

	const statusLabel: Record<OrderStatus, string> = {
		pending: 'Menunggu',
		confirmed: 'Dikonfirmasi',
		preparing: 'Diproses',
		ready: 'Siap',
		completed: 'Selesai',
		cancelled: 'Dibatalkan'
	};

	const statusColor: Record<OrderStatus, string> = {
		pending: 'bg-yellow-50 text-yellow-700',
		confirmed: 'bg-blue-50 text-blue-700',
		preparing: 'bg-orange-50 text-orange-700',
		ready: 'bg-green-50 text-green-700',
		completed: 'bg-gray-100 text-gray-500',
		cancelled: 'bg-red-50 text-red-400'
	};

	const nextStatus: Partial<Record<OrderStatus, OrderStatus>> = {
		pending: 'confirmed',
		confirmed: 'preparing',
		preparing: 'ready',
		ready: 'completed'
	};

	const nextStatusLabel: Partial<Record<OrderStatus, string>> = {
		pending: 'Konfirmasi',
		confirmed: 'Mulai Proses',
		preparing: 'Tandai Siap',
		ready: 'Selesaikan'
	};

	function getBranchName(id: number) {
		return data.branches.find((b) => b.id === id)?.name ?? `Cabang #${id}`;
	}

	function formatDate(d: string) {
		return new Date(d).toLocaleString('id-ID', {
			day: '2-digit',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

<svelte:head>
	<title>Pesanan — RestoQRIS</title>
</svelte:head>

<div class="max-w-5xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h2 class="text-2xl font-bold text-gray-800">Manajemen Pesanan</h2>
			<p class="mt-1 text-gray-500 text-sm">Monitor dan kelola pesanan aktif.</p>
		</div>
		<a
			href="/dashboard/pos"
			class="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg"
		>
			+ Pesanan Baru
		</a>
	</div>

	<!-- Feedback -->
	{#if form && !form.success && 'error' in form && form.error}
		<div class="mb-4 rounded-lg bg-red-50 border border-red-200 text-red-700 px-4 py-3 text-sm">
			{form.error}
		</div>
	{/if}
	{#if form?.success}
		<div class="mb-4 rounded-lg bg-green-50 border border-green-200 text-green-700 px-4 py-3 text-sm">
			Status pesanan berhasil diperbarui.
		</div>
	{/if}

	<!-- Filters -->
	<div class="bg-white rounded-xl shadow-sm p-4 mb-6 flex flex-wrap gap-3 items-end">
		<div class="flex-1 min-w-[160px]">
			<label for="filter-branch" class="block text-xs font-medium text-gray-500 mb-1">Cabang</label>
			<select
				id="filter-branch"
				bind:value={filterBranch}
				class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
			>
				<option value={null}>Semua Cabang</option>
				{#each data.branches as b}
					<option value={b.id}>{b.name}</option>
				{/each}
			</select>
		</div>
		<div class="flex-1 min-w-[140px]">
			<label for="filter-status" class="block text-xs font-medium text-gray-500 mb-1">Status</label>
			<select
				id="filter-status"
				bind:value={filterStatus}
				class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
			>
				<option value="">Semua Status</option>
				<option value="pending">Menunggu</option>
				<option value="confirmed">Dikonfirmasi</option>
				<option value="preparing">Diproses</option>
				<option value="ready">Siap</option>
				<option value="completed">Selesai</option>
				<option value="cancelled">Dibatalkan</option>
			</select>
		</div>
		<button
			onclick={applyFilters}
			class="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg shrink-0"
		>
			Terapkan Filter
		</button>
	</div>

	<!-- Orders list -->
	{#if data.orders.length === 0}
		<div class="bg-white rounded-xl shadow-sm p-12 text-center">
			<span class="text-5xl mb-4 block">📋</span>
			<p class="text-gray-500">Tidak ada pesanan ditemukan.</p>
		</div>
	{:else}
		<div class="space-y-3">
			{#each data.orders as order (order.id)}
				{@const isExpanded = expandedOrderId === order.id}
				<div class="bg-white rounded-xl shadow-sm overflow-hidden">
					<!-- Order header -->
					<div class="px-5 py-4 flex items-start justify-between gap-4">
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2 flex-wrap">
								<p class="font-bold text-gray-800">{order.order_code}</p>
								<span class="text-xs px-2 py-0.5 rounded-full font-medium {statusColor[order.status]}">
									{statusLabel[order.status]}
								</span>
								<span class="text-xs text-gray-400">
									{order.order_type === 'dine_in' ? '🪑 Dine In' : '🥡 Takeaway'}
								</span>
							</div>
							<p class="text-xs text-gray-400 mt-1">
								{getBranchName(order.branch_id)} · {formatDate(order.created_at)}
								{#if order.table_id}· Meja #{order.table_id}{/if}
							</p>
						</div>
						<div class="flex items-center gap-2 shrink-0">
							<span class="font-bold text-orange-500 text-sm">{formatPrice(order.total_amount)}</span>

							<!-- Next status button -->
							{#if nextStatus[order.status]}
								<form
									method="POST"
									action="?/updateStatus"
									use:enhance={() => {
										return async ({ update }) => { await update(); };
									}}
								>
									<input type="hidden" name="order_id" value={order.id} />
									<input type="hidden" name="status" value={nextStatus[order.status]} />
									<button
										type="submit"
										class="px-3 py-1.5 bg-orange-500 hover:bg-orange-600 text-white text-xs font-semibold rounded-lg"
									>
										{nextStatusLabel[order.status]}
									</button>
								</form>
							{/if}

							<!-- Cancel button (only for pending/confirmed) -->
							{#if order.status === 'pending' || order.status === 'confirmed'}
								<form
									method="POST"
									action="?/updateStatus"
									use:enhance={() => {
										return async ({ update }) => { await update(); };
									}}
								>
									<input type="hidden" name="order_id" value={order.id} />
									<input type="hidden" name="status" value="cancelled" />
									<button
										type="submit"
										class="px-3 py-1.5 bg-gray-100 hover:bg-red-100 text-gray-500 hover:text-red-600 text-xs font-semibold rounded-lg"
									>
										Batalkan
									</button>
								</form>
							{/if}

							<!-- Toggle detail -->
							<button
								onclick={() => (expandedOrderId = isExpanded ? null : order.id)}
								class="text-gray-400 hover:text-gray-600 text-sm"
								aria-label="Toggle detail"
							>
								{isExpanded ? '▲' : '▼'}
							</button>
						</div>
					</div>

					<!-- Order detail (expandable) -->
					{#if isExpanded}
						<div class="border-t border-gray-100 px-5 py-4 bg-gray-50">
							{#if order.notes}
								<p class="text-xs text-gray-500 mb-3 italic">Catatan: {order.notes}</p>
							{/if}
							<table class="w-full text-sm">
								<thead>
									<tr class="text-xs text-gray-400 text-left">
										<th class="pb-2 font-medium">Item</th>
										<th class="pb-2 font-medium text-center">Qty</th>
										<th class="pb-2 font-medium text-right">Harga</th>
										<th class="pb-2 font-medium text-right">Subtotal</th>
									</tr>
								</thead>
								<tbody class="divide-y divide-gray-100">
									{#each order.items as item}
										<tr>
											<td class="py-2 text-gray-700">
												{item.item_name}
												{#if item.notes}
													<span class="text-xs text-gray-400 italic"> – {item.notes}</span>
												{/if}
											</td>
											<td class="py-2 text-center text-gray-500">{item.quantity}</td>
											<td class="py-2 text-right text-gray-500">{formatPrice(item.unit_price)}</td>
											<td class="py-2 text-right font-medium text-gray-700">{formatPrice(item.subtotal)}</td>
										</tr>
									{/each}
								</tbody>
								<tfoot>
									<tr class="border-t border-gray-200">
										<td colspan="3" class="pt-2 text-right text-xs text-gray-500 font-medium">Subtotal</td>
										<td class="pt-2 text-right text-sm font-medium text-gray-700">{formatPrice(order.subtotal)}</td>
									</tr>
									<tr>
										<td colspan="3" class="text-right text-xs font-bold text-gray-800">Total</td>
										<td class="text-right text-sm font-bold text-orange-500">{formatPrice(order.total_amount)}</td>
									</tr>
								</tfoot>
							</table>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
