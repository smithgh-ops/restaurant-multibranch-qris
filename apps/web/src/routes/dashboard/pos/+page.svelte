<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData, ActionData } from './$types';
	import type { MenuItem, Payment, RestaurantTable } from '$lib/api/client';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	// ── State ─────────────────────────────────────────────────────────────────
	let selectedBranchId = $state<number | null>(null);
	let selectedOrderType = $state<'dine_in' | 'takeaway' | 'delivery'>('dine_in');
	let selectedTableId = $state<number | null>(null);
	let tables = $state<RestaurantTable[]>([]);
	let loadingTables = $state(false);
	let orderNotes = $state('');
	let selectedCategoryId = $state<number | null>(null);
	let orderSuccess = $state<string | null>(null);
	let paymentInfo = $state<Payment | null>(null);
	let paymentWarning = $state<string | null>(null);

	interface CartItem {
		menuItemId: number;
		name: string;
		unitPrice: number;
		quantity: number;
		notes: string;
	}
	let cart = $state<CartItem[]>([]);

	// ── Derived ───────────────────────────────────────────────────────────────
	const filteredItems = $derived(
		selectedCategoryId
			? data.items.filter((item) => item.category_id === selectedCategoryId)
			: data.items
	);

	const cartTotal = $derived(cart.reduce((sum, item) => sum + item.unitPrice * item.quantity, 0));

	const cartJson = $derived(
		JSON.stringify(
			cart.map((item) => ({
				menu_item_id: item.menuItemId,
				quantity: item.quantity,
				notes: item.notes || undefined
			}))
		)
	);

	const priceFormatter = new Intl.NumberFormat('id-ID', {
		style: 'currency',
		currency: 'IDR',
		minimumFractionDigits: 0
	});

	function formatPrice(price: number | string) {
		return priceFormatter.format(Number(price));
	}

	// ── Cart helpers ──────────────────────────────────────────────────────────
	function addToCart(item: MenuItem) {
		const idx = cart.findIndex((c) => c.menuItemId === item.id);
		if (idx >= 0) {
			cart[idx].quantity += 1;
		} else {
			cart.push({
				menuItemId: item.id,
				name: item.name,
				unitPrice: Number(item.base_price),
				quantity: 1,
				notes: ''
			});
		}
	}

	function incrementCartItem(menuItemId: number) {
		const idx = cart.findIndex((c) => c.menuItemId === menuItemId);
		if (idx >= 0) {
			cart[idx].quantity += 1;
		}
	}

	function removeFromCart(menuItemId: number) {
		const idx = cart.findIndex((c) => c.menuItemId === menuItemId);
		if (idx >= 0) {
			if (cart[idx].quantity > 1) {
				cart[idx].quantity -= 1;
			} else {
				cart.splice(idx, 1);
			}
		}
	}

	function clearCart() {
		cart = [];
		selectedTableId = null;
		orderNotes = '';
		orderSuccess = null;
		paymentInfo = null;
		paymentWarning = null;
	}

	// ── Branch / table loading ────────────────────────────────────────────────
	function onBranchChange() {
		tables = [];
		selectedTableId = null;
	}

	// When form action returns tables list, update local state
	$effect(() => {
		if (form && 'tables' in form && Array.isArray(form.tables)) {
			tables = form.tables as RestaurantTable[];
			loadingTables = false;
		}
		if (form && 'orderCode' in form && form.orderCode) {
			orderSuccess = form.orderCode as string;
			paymentInfo = ('payment' in form ? (form.payment as Payment) : null) ?? null;
			paymentWarning = 'paymentWarning' in form ? (form.paymentWarning as string) : null;
			cart = [];
			selectedTableId = null;
			orderNotes = '';
		}
	});
</script>

<svelte:head>
	<title>POS Kasir — RestoQRIS</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
	<div class="mb-6">
		<h2 class="text-2xl font-bold text-gray-800">POS Kasir</h2>
		<p class="mt-1 text-gray-500 text-sm">Buat pesanan baru untuk pelanggan.</p>
	</div>

	<!-- Order success banner -->
	{#if orderSuccess}
		<div class="mb-4 rounded-lg bg-green-50 border border-green-200 text-green-700 px-4 py-3 text-sm flex items-center justify-between">
			<span>✅ Pesanan berhasil dibuat! Kode: <strong>{orderSuccess}</strong></span>
			<button onclick={clearCart} class="text-green-700 hover:text-green-900 font-medium">Buat Pesanan Baru</button>
		</div>
		{#if paymentWarning}
			<div class="mb-4 rounded-lg bg-yellow-50 border border-yellow-200 text-yellow-700 px-4 py-3 text-sm">
				⚠️ Pesanan tercatat, tetapi invoice QRIS belum dibuat: {paymentWarning}
			</div>
		{/if}
		{#if paymentInfo}
			<div class="mb-4 rounded-xl border border-orange-200 bg-orange-50 p-4">
				<h3 class="text-sm font-semibold text-orange-700 mb-2">Pembayaran QRIS</h3>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4 items-start">
					<div class="space-y-1 text-sm text-gray-700">
						<p>Kode invoice: <strong>{paymentInfo.gateway_invoice_id}</strong></p>
						<p>Total: <strong>{formatPrice(paymentInfo.amount)}</strong></p>
						<p>Status: <strong class="uppercase">{paymentInfo.status}</strong></p>
						{#if paymentInfo.expiry_at}
							<p>Kadaluarsa: <strong>{new Date(paymentInfo.expiry_at).toLocaleString('id-ID')}</strong></p>
						{/if}
					</div>
					{#if paymentInfo.qr_code_url}
						<img src={paymentInfo.qr_code_url} alt="QRIS Dinamis" class="w-48 h-48 rounded-lg border border-orange-200 bg-white p-2" />
					{:else if paymentInfo.qr_string}
						<pre class="text-xs bg-white border border-orange-200 rounded-lg p-2 overflow-x-auto">{paymentInfo.qr_string}</pre>
					{/if}
				</div>
			</div>
		{/if}
	{/if}

	<!-- Error banner -->
	{#if form && !form.success && 'error' in form && form.error && !('tables' in form)}
		<div class="mb-4 rounded-lg bg-red-50 border border-red-200 text-red-700 px-4 py-3 text-sm">
			{form.error}
		</div>
	{/if}

	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- LEFT: Menu ─────────────────────────────────────────────────────── -->
		<div class="lg:col-span-2 space-y-4">
			<!-- Branch selector -->
			<div class="bg-white rounded-xl shadow-sm p-4 flex flex-wrap gap-4 items-end">
				<div class="flex-1 min-w-[160px]">
					<label for="branch-select" class="block text-xs font-medium text-gray-500 mb-1">Cabang</label>
					<select
						id="branch-select"
						bind:value={selectedBranchId}
						onchange={onBranchChange}
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
					>
						<option value={null}>Pilih cabang</option>
						{#each data.branches.filter((b) => b.is_active) as branch}
							<option value={branch.id}>{branch.name}</option>
						{/each}
					</select>
				</div>

				<div class="flex-1 min-w-[140px]">
					<label for="order-type" class="block text-xs font-medium text-gray-500 mb-1">Tipe Pesanan</label>
					<select
						id="order-type"
						bind:value={selectedOrderType}
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
					>
						<option value="dine_in">Makan di Tempat</option>
						<option value="takeaway">Dibawa Pulang</option>
						<option value="delivery">Delivery</option>
					</select>
				</div>

				{#if selectedOrderType === 'dine_in' && selectedBranchId}
					<div class="flex-1 min-w-[140px]">
						<label for="table-select" class="block text-xs font-medium text-gray-500 mb-1">Meja</label>
						{#if tables.length === 0}
							<!-- Load tables via server action -->
							<form
								method="POST"
								action="?/loadTables"
								use:enhance={() => {
									loadingTables = true;
									return async ({ update }) => { await update({ reset: false }); };
								}}
							>
								<input type="hidden" name="branch_id" value={selectedBranchId} />
								<button
									type="submit"
									class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-500 hover:bg-gray-50 text-left"
								>
									{loadingTables ? 'Memuat...' : 'Muat daftar meja →'}
								</button>
							</form>
						{:else}
							<select
								id="table-select"
								bind:value={selectedTableId}
								class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
							>
								<option value={null}>Pilih meja</option>
								{#each tables.filter((t) => t.is_active) as t}
									<option value={t.id}>Meja {t.table_number} (kap. {t.capacity})</option>
								{/each}
							</select>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Category filter -->
			<div class="flex gap-2 overflow-x-auto pb-1">
				<button
					onclick={() => (selectedCategoryId = null)}
					class="shrink-0 px-4 py-1.5 rounded-full text-sm font-medium transition-colors
					{selectedCategoryId === null ? 'bg-orange-500 text-white' : 'bg-white text-gray-600 border border-gray-200 hover:bg-gray-50'}"
				>
					Semua
				</button>
				{#each data.categories.filter((c) => c.is_active) as cat}
					<button
						onclick={() => (selectedCategoryId = cat.id)}
						class="shrink-0 px-4 py-1.5 rounded-full text-sm font-medium transition-colors
						{selectedCategoryId === cat.id ? 'bg-orange-500 text-white' : 'bg-white text-gray-600 border border-gray-200 hover:bg-gray-50'}"
					>
						{cat.name}
					</button>
				{/each}
			</div>

			<!-- Menu grid -->
			{#if filteredItems.length === 0}
				<div class="bg-white rounded-xl shadow-sm p-12 text-center text-gray-400">
					Tidak ada item menu.
				</div>
			{:else}
				<div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
					{#each filteredItems as item}
						<button
							onclick={() => addToCart(item)}
							class="bg-white rounded-xl shadow-sm p-4 text-left hover:shadow-md hover:ring-2 hover:ring-orange-300 transition-all"
						>
							<div class="w-full h-24 rounded-lg bg-orange-50 flex items-center justify-center text-4xl mb-3">
								🍜
							</div>
							<p class="font-semibold text-gray-800 text-sm line-clamp-1">{item.name}</p>
							<p class="text-orange-500 font-semibold text-sm mt-1">{formatPrice(item.base_price)}</p>
						</button>
					{/each}
				</div>
			{/if}
		</div>

		<!-- RIGHT: Cart ─────────────────────────────────────────────────────── -->
		<div class="lg:col-span-1">
			<div class="bg-white rounded-xl shadow-sm p-5 sticky top-6">
				<h3 class="font-bold text-gray-800 mb-4 flex items-center gap-2">
					🛒 Keranjang
					{#if cart.length > 0}
						<span class="bg-orange-100 text-orange-600 text-xs font-semibold px-2 py-0.5 rounded-full">
							{cart.reduce((s, i) => s + i.quantity, 0)} item
						</span>
					{/if}
				</h3>

				{#if cart.length === 0}
					<div class="py-8 text-center text-gray-400 text-sm">
						Belum ada item ditambahkan.
					</div>
				{:else}
					<div class="space-y-3 mb-4 max-h-64 overflow-y-auto">
						{#each cart as item}
							<div class="flex items-start justify-between gap-2">
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium text-gray-800 truncate">{item.name}</p>
									<p class="text-xs text-gray-400">{formatPrice(item.unitPrice)} × {item.quantity}</p>
								</div>
								<div class="flex items-center gap-1 shrink-0">
									<button
										onclick={() => removeFromCart(item.menuItemId)}
										class="w-6 h-6 rounded-full bg-gray-100 text-gray-600 hover:bg-red-100 hover:text-red-600 text-sm font-bold flex items-center justify-center"
									>−</button>
									<span class="w-5 text-center text-sm font-semibold">{item.quantity}</span>
									<button
										onclick={() => incrementCartItem(item.menuItemId)}
										class="w-6 h-6 rounded-full bg-gray-100 text-gray-600 hover:bg-orange-100 hover:text-orange-600 text-sm font-bold flex items-center justify-center"
									>+</button>
								</div>
							</div>
						{/each}
					</div>

					<div class="border-t border-gray-100 pt-3 mb-4">
						<div class="flex justify-between font-bold text-gray-800">
							<span>Total</span>
							<span class="text-orange-500">{formatPrice(cartTotal)}</span>
						</div>
					</div>

					<div class="mb-3">
						<label for="order-notes" class="block text-xs font-medium text-gray-500 mb-1">Catatan Pesanan</label>
						<textarea
							id="order-notes"
							bind:value={orderNotes}
							rows="2"
							placeholder="Catatan khusus (opsional)"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400 resize-none"
						></textarea>
					</div>

					{#if selectedBranchId}
						<form
							method="POST"
							action="?/checkout"
							use:enhance={() => {
								return async ({ update }) => { await update({ reset: false }); };
							}}
						>
							<input type="hidden" name="branch_id" value={selectedBranchId} />
							<input type="hidden" name="order_type" value={selectedOrderType} />
							{#if selectedTableId}
								<input type="hidden" name="table_id" value={selectedTableId} />
							{/if}
							<input type="hidden" name="notes" value={orderNotes} />
							<input type="hidden" name="items" value={cartJson} />
							<button
								type="submit"
								class="w-full py-3 bg-orange-500 hover:bg-orange-600 text-white font-bold rounded-lg transition-colors"
							>
								Checkout →
							</button>
						</form>
					{:else}
						<p class="text-xs text-center text-gray-400">Pilih cabang untuk checkout.</p>
					{/if}

					<button
						onclick={clearCart}
						class="w-full mt-2 py-2 text-sm text-gray-400 hover:text-red-500 transition-colors"
					>
						Kosongkan Keranjang
					</button>
				{/if}
			</div>
		</div>
	</div>
</div>
