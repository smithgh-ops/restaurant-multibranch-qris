<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData, ActionData } from './$types';
	import type { PublicMenuItem, SelfOrderResponse } from '$lib/api/client';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	interface CartItem {
		menuItemId: number;
		name: string;
		price: number;
		quantity: number;
		notes: string;
	}
	let cart = $state<CartItem[]>([]);
	let orderNotes = $state('');
	let selectedCategoryId = $state<number | null>(null);
	let successOrder = $state<SelfOrderResponse | null>(null);

	$effect(() => {
		if (form && 'order' in form && form.order) {
			successOrder = form.order as SelfOrderResponse;
			cart = [];
			orderNotes = '';
		}
	});

	const priceFormatter = new Intl.NumberFormat('id-ID', {
		style: 'currency',
		currency: 'IDR',
		minimumFractionDigits: 0
	});

	function formatPrice(p: number | string) {
		return priceFormatter.format(Number(p));
	}

	const visibleItems = $derived(
		selectedCategoryId
			? data.menu.categories.find((c) => c.id === selectedCategoryId)?.items ?? []
			: data.menu.categories.flatMap((c) => c.items)
	);

	const cartTotal = $derived(cart.reduce((s, i) => s + i.price * i.quantity, 0));

	const cartJson = $derived(
		JSON.stringify(
			cart.map((i) => ({
				menu_item_id: i.menuItemId,
				quantity: i.quantity,
				notes: i.notes || undefined
			}))
		)
	);

	function incrementCartItem(menuItemId: number, itemName: string, itemPrice: number) {
		const idx = cart.findIndex((c) => c.menuItemId === menuItemId);
		if (idx >= 0) {
			cart[idx].quantity += 1;
		} else {
			cart.push({ menuItemId, name: itemName, price: itemPrice, quantity: 1, notes: '' });
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

	function resetOrder() {
		successOrder = null;
	}
</script>

<svelte:head>
	<title>Pesan — {data.menu.branch.name}, Meja {data.menu.table.table_number}</title>
</svelte:head>

<!-- Minimal public layout (no dashboard chrome) -->
<div class="min-h-screen bg-gray-50">
	<!-- Header -->
	<header class="bg-white shadow-sm sticky top-0 z-10">
		<div class="max-w-lg mx-auto px-4 py-3 flex items-center justify-between">
			<div>
				<p class="font-bold text-orange-600 text-lg">RestoQRIS</p>
				<p class="text-xs text-gray-500">{data.menu.branch.name} · Meja {data.menu.table.table_number}</p>
			</div>
			{#if cart.length > 0 && !successOrder}
				<span class="bg-orange-500 text-white text-xs font-semibold px-2 py-0.5 rounded-full">
					{cart.reduce((s, i) => s + i.quantity, 0)} item
				</span>
			{/if}
		</div>
	</header>

	<div class="max-w-lg mx-auto px-4 py-4 pb-40">
		<!-- Success state -->
		{#if successOrder}
			<div class="mt-8 bg-white rounded-2xl shadow-sm p-8 text-center">
				<span class="text-5xl block mb-4">🎉</span>
				<h2 class="text-xl font-bold text-gray-800 mb-2">Pesanan Diterima!</h2>
				<p class="text-gray-500 text-sm mb-1">Kode pesanan Anda:</p>
				<p class="text-2xl font-bold text-orange-600 mb-4">{successOrder.order_code}</p>
				<p class="text-sm text-gray-500 mb-6">
					Total: <strong>{formatPrice(successOrder.total_amount)}</strong>
				</p>
				<p class="text-xs text-gray-400 mb-6">Staf kami akan segera memproses pesanan Anda.</p>
				<button
					onclick={resetOrder}
					class="w-full py-3 bg-orange-500 hover:bg-orange-600 text-white font-semibold rounded-xl"
				>
					Pesan Lagi
				</button>
			</div>
		{:else}
			<!-- Error -->
			{#if form && !form.success && 'error' in form && form.error}
				<div class="mb-4 rounded-lg bg-red-50 border border-red-200 text-red-700 px-4 py-3 text-sm">
					{form.error}
				</div>
			{/if}

			<!-- Category tabs -->
			<div class="flex gap-2 overflow-x-auto pb-2 mb-4">
				<button
					onclick={() => (selectedCategoryId = null)}
					class="shrink-0 px-4 py-1.5 rounded-full text-sm font-medium transition-colors
					{selectedCategoryId === null ? 'bg-orange-500 text-white' : 'bg-white text-gray-600 border border-gray-200'}"
				>
					Semua
				</button>
				{#each data.menu.categories as cat}
					<button
						onclick={() => (selectedCategoryId = cat.id)}
						class="shrink-0 px-4 py-1.5 rounded-full text-sm font-medium transition-colors
						{selectedCategoryId === cat.id ? 'bg-orange-500 text-white' : 'bg-white text-gray-600 border border-gray-200'}"
					>
						{cat.name}
					</button>
				{/each}
			</div>

			<!-- Menu items -->
			<div class="grid grid-cols-2 gap-3 mb-6">
				{#each visibleItems as item}
					{@const inCart = cart.find((c) => c.menuItemId === item.id)}
					<button
						onclick={() => incrementCartItem(item.id, item.name, Number(item.price))}
						class="bg-white rounded-xl shadow-sm p-3 text-left hover:shadow-md hover:ring-2 hover:ring-orange-300 transition-all"
					>
						{#if item.image_url}
							<img
								src={item.image_url}
								alt={item.name}
								class="w-full h-24 object-cover rounded-lg mb-2"
								loading="lazy"
							/>
						{:else}
							<div class="w-full h-24 rounded-lg bg-orange-50 flex items-center justify-center text-3xl mb-2">
								🍜
							</div>
						{/if}
						<p class="font-semibold text-gray-800 text-sm line-clamp-1">{item.name}</p>
						<div class="flex items-center justify-between mt-1">
							<p class="text-orange-500 font-semibold text-sm">{formatPrice(item.price)}</p>
							{#if inCart}
								<span class="bg-orange-100 text-orange-600 text-xs font-bold px-2 py-0.5 rounded-full">
									×{inCart.quantity}
								</span>
							{/if}
						</div>
					</button>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Sticky cart + checkout -->
	{#if cart.length > 0 && !successOrder}
		<div class="fixed bottom-0 left-0 right-0 z-20 bg-white border-t border-gray-200 shadow-lg">
			<div class="max-w-lg mx-auto px-4 py-3">
				<!-- Cart items summary -->
				<div class="max-h-40 overflow-y-auto mb-3 space-y-1">
					{#each cart as item}
						<div class="flex items-center justify-between text-sm">
							<span class="text-gray-700 truncate flex-1">{item.name}</span>
							<div class="flex items-center gap-2 ml-2 shrink-0">
								<button
									onclick={() => removeFromCart(item.menuItemId)}
									class="w-6 h-6 rounded-full bg-gray-100 hover:bg-gray-200 text-gray-700 flex items-center justify-center text-xs font-bold"
								>
									−
								</button>
								<span class="w-5 text-center font-semibold">{item.quantity}</span>
								<button
									onclick={() => incrementCartItem(item.menuItemId, item.name, item.price)}
									class="w-6 h-6 rounded-full bg-orange-100 hover:bg-orange-200 text-orange-600 flex items-center justify-center text-xs font-bold"
								>
									+
								</button>
								<span class="text-gray-500 w-20 text-right">{formatPrice(item.price * item.quantity)}</span>
							</div>
						</div>
					{/each}
				</div>

				<!-- Checkout form -->
				<form method="POST" action="?/order" use:enhance={() => async ({ update }) => { await update({ reset: false }); }}>
					<input type="hidden" name="items" value={cartJson} />
					<textarea
						name="notes"
						bind:value={orderNotes}
						placeholder="Catatan pesanan (opsional)"
						rows="1"
						class="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm resize-none focus:outline-none focus:ring-2 focus:ring-orange-400 mb-3"
					></textarea>
					<button
						type="submit"
						class="w-full py-3 bg-orange-500 hover:bg-orange-600 text-white font-bold rounded-xl flex items-center justify-center gap-2"
					>
						<span>Pesan Sekarang</span>
						<span class="bg-white/20 px-2 py-0.5 rounded-lg text-sm">{formatPrice(cartTotal)}</span>
					</button>
				</form>
			</div>
		</div>
	{/if}
</div>
