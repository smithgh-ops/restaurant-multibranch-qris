<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData, ActionData } from './$types';
	import type { MenuItem } from '$lib/api/client';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let activeTab = $state<'categories' | 'items'>('items');
	let showCategoryForm = $state(false);
	let showItemForm = $state(false);
	let editingItem = $state<MenuItem | null>(null);
	let selectedCategoryFilter = $state<number | null>(null);

	const filteredItems = $derived(
		selectedCategoryFilter
			? data.items.filter((item) => item.category_id === selectedCategoryFilter)
			: data.items
	);

	const priceFormatter = new Intl.NumberFormat('id-ID', {
		style: 'currency',
		currency: 'IDR',
		minimumFractionDigits: 0
	});

	function formatPrice(price: string) {
		const num = parseFloat(price);
		return priceFormatter.format(num);
	}

	function getCategoryName(id: number) {
		return data.categories.find((c) => c.id === id)?.name ?? '-';
	}
</script>

<svelte:head>
	<title>Manajemen Menu — RestoQRIS</title>
</svelte:head>

<div class="max-w-5xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h2 class="text-2xl font-bold text-gray-800">Manajemen Menu</h2>
			<p class="mt-1 text-gray-500 text-sm">Kelola kategori dan item menu restoran Anda.</p>
		</div>
	</div>

	<!-- Feedback -->
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

	<!-- Tabs -->
	<div class="flex gap-2 mb-6 border-b border-gray-200">
		<button
			onclick={() => (activeTab = 'items')}
			class="px-4 py-2 text-sm font-medium border-b-2 transition-colors {activeTab === 'items'
				? 'border-orange-500 text-orange-600'
				: 'border-transparent text-gray-500 hover:text-gray-700'}"
		>
			Item Menu ({data.items.length})
		</button>
		<button
			onclick={() => (activeTab = 'categories')}
			class="px-4 py-2 text-sm font-medium border-b-2 transition-colors {activeTab === 'categories'
				? 'border-orange-500 text-orange-600'
				: 'border-transparent text-gray-500 hover:text-gray-700'}"
		>
			Kategori ({data.categories.length})
		</button>
	</div>

	<!-- CATEGORIES TAB -->
	{#if activeTab === 'categories'}
		<div class="flex justify-end mb-4">
			<button
				onclick={() => (showCategoryForm = !showCategoryForm)}
				class="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg"
			>
				{showCategoryForm ? 'Batal' : '+ Kategori Baru'}
			</button>
		</div>

		{#if showCategoryForm}
			<div class="bg-white rounded-xl shadow-sm p-6 mb-4">
				<h3 class="font-semibold text-gray-800 mb-4">Tambah Kategori</h3>
				<form
					method="POST"
					action="?/createCategory"
					use:enhance={() => {
						return async ({ update }) => {
							await update();
							if (form?.success) showCategoryForm = false;
						};
					}}
					class="flex flex-col sm:flex-row gap-3"
				>
					<input
						name="name"
						type="text"
						required
						placeholder="Nama kategori"
						class="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
					/>
					<input
						name="description"
						type="text"
						placeholder="Deskripsi (opsional)"
						class="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
					/>
					<button
						type="submit"
						class="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg shrink-0"
					>
						Simpan
					</button>
				</form>
			</div>
		{/if}

		{#if data.categories.length === 0}
			<div class="bg-white rounded-xl shadow-sm p-12 text-center">
				<span class="text-5xl mb-4 block">🗂️</span>
				<p class="text-gray-500">Belum ada kategori. Buat kategori pertama Anda.</p>
			</div>
		{:else}
			<div class="space-y-2">
				{#each data.categories as cat}
					<div class="bg-white rounded-xl shadow-sm px-5 py-4 flex items-center justify-between">
						<div>
							<p class="font-semibold text-gray-800">{cat.name}</p>
							{#if cat.description}
								<p class="text-xs text-gray-500 mt-0.5">{cat.description}</p>
							{/if}
						</div>
						<span
							class="text-xs px-2 py-0.5 rounded-full {cat.is_active
								? 'bg-green-50 text-green-600'
								: 'bg-gray-100 text-gray-400'}"
						>
							{cat.is_active ? 'Aktif' : 'Nonaktif'}
						</span>
					</div>
				{/each}
			</div>
		{/if}
	{/if}

	<!-- ITEMS TAB -->
	{#if activeTab === 'items'}
		<div class="flex flex-col sm:flex-row gap-3 justify-between mb-4">
			<select
				bind:value={selectedCategoryFilter}
				class="rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
			>
				<option value={null}>Semua Kategori</option>
				{#each data.categories as cat}
					<option value={cat.id}>{cat.name}</option>
				{/each}
			</select>
			<button
				onclick={() => {
					showItemForm = !showItemForm;
					editingItem = null;
				}}
				class="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold rounded-lg"
			>
				{showItemForm ? 'Batal' : '+ Item Baru'}
			</button>
		</div>

		{#if showItemForm}
			<div class="bg-white rounded-xl shadow-sm p-6 mb-4">
				<h3 class="font-semibold text-gray-800 mb-4">Tambah Item Menu</h3>
				<form
					method="POST"
					action="?/createItem"
					use:enhance={() => {
						return async ({ update }) => {
							await update();
							if (form?.success) showItemForm = false;
						};
					}}
					class="space-y-4"
				>
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<div>
							<label for="new-item-category" class="block text-sm font-medium text-gray-700 mb-1">Kategori</label>
							<select
								id="new-item-category"
								name="category_id"
								required
								class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
							>
								<option value="">Pilih kategori</option>
								{#each data.categories as cat}
									<option value={cat.id}>{cat.name}</option>
								{/each}
							</select>
						</div>
						<div>
							<label for="new-item-name" class="block text-sm font-medium text-gray-700 mb-1">Nama Item</label>
							<input
								id="new-item-name"
								name="name"
								type="text"
								required
								placeholder="Nama menu"
								class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
							/>
						</div>
						<div>
							<label for="new-item-price" class="block text-sm font-medium text-gray-700 mb-1">Harga Dasar (IDR)</label>
							<input
								id="new-item-price"
								name="base_price"
								type="number"
								required
								min="0"
								step="100"
								placeholder="25000"
								class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
							/>
						</div>
						<div>
							<label for="new-item-desc" class="block text-sm font-medium text-gray-700 mb-1">Deskripsi</label>
							<input
								id="new-item-desc"
								name="description"
								type="text"
								placeholder="Deskripsi singkat (opsional)"
								class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
							/>
						</div>
					</div>
					<div>
						<label for="new-item-image" class="block text-sm font-medium text-gray-700 mb-1">URL Gambar</label>
						<input
							id="new-item-image"
							name="image_url"
							type="url"
							placeholder="https://... (opsional)"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
						/>
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
							onclick={() => (showItemForm = false)}
							class="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 text-sm font-semibold rounded-lg"
						>
							Batal
						</button>
					</div>
				</form>
			</div>
		{/if}

		<!-- Edit item form -->
		{#if editingItem}
			<div class="bg-white rounded-xl shadow-sm p-6 mb-4">
				<h3 class="font-semibold text-gray-800 mb-4">Edit: {editingItem.name}</h3>
				<form
					method="POST"
					action="?/updateItem"
					use:enhance={() => {
						return async ({ update }) => {
							await update();
							if (form?.success) editingItem = null;
						};
					}}
					class="space-y-4"
				>
					<input type="hidden" name="id" value={editingItem.id} />
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<div>
							<label for="edit-item-name" class="block text-sm font-medium text-gray-700 mb-1">Nama Item</label>
							<input
								id="edit-item-name"
								name="name"
								type="text"
								value={editingItem.name}
								class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
							/>
						</div>
						<div>
							<label for="edit-item-price" class="block text-sm font-medium text-gray-700 mb-1">Harga Dasar (IDR)</label>
							<input
								id="edit-item-price"
								name="base_price"
								type="number"
								value={editingItem.base_price}
								min="0"
								step="100"
								class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
							/>
						</div>
						<div>
							<label for="edit-item-desc" class="block text-sm font-medium text-gray-700 mb-1">Deskripsi</label>
							<input
								id="edit-item-desc"
								name="description"
								type="text"
								value={editingItem.description ?? ''}
								class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
							/>
						</div>
						<div>
							<label for="edit-item-status" class="block text-sm font-medium text-gray-700 mb-1">Status</label>
							<select
								id="edit-item-status"
								name="is_active"
								class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
							>
								<option value="true" selected={editingItem.is_active}>Aktif</option>
								<option value="false" selected={!editingItem.is_active}>Nonaktif</option>
							</select>
						</div>
					</div>
					<div>
						<label for="edit-item-image" class="block text-sm font-medium text-gray-700 mb-1">URL Gambar</label>
						<div class="flex gap-3 items-start">
							{#if editingItem.image_url}
								<img src={editingItem.image_url} alt={editingItem.name} class="w-14 h-14 rounded-lg object-cover border border-gray-200 shrink-0" />
							{/if}
							<input
								id="edit-item-image"
								name="image_url"
								type="url"
								value={editingItem.image_url ?? ''}
								placeholder="https://... (opsional)"
								class="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
							/>
						</div>
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
							onclick={() => (editingItem = null)}
							class="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 text-sm font-semibold rounded-lg"
						>
							Batal
						</button>
					</div>
				</form>
			</div>
		{/if}

		{#if filteredItems.length === 0}
			<div class="bg-white rounded-xl shadow-sm p-12 text-center">
				<span class="text-5xl mb-4 block">🍽️</span>
				<p class="text-gray-500">Belum ada item menu. Tambahkan item pertama Anda.</p>
			</div>
		{:else}
			<div class="space-y-2">
				{#each filteredItems as item}
					<div class="bg-white rounded-xl shadow-sm px-5 py-4 flex items-center justify-between gap-4">
						<div class="flex items-center gap-4">
							{#if item.image_url}
								<img src={item.image_url} alt={item.name} class="w-12 h-12 rounded-lg object-cover" />
							{:else}
								<div class="w-12 h-12 rounded-lg bg-orange-50 flex items-center justify-center text-2xl">
									🍜
								</div>
							{/if}
							<div>
								<p class="font-semibold text-gray-800">{item.name}</p>
								<p class="text-xs text-gray-400">{getCategoryName(item.category_id)}</p>
								{#if item.description}
									<p class="text-xs text-gray-500 mt-0.5 line-clamp-1">{item.description}</p>
								{/if}
							</div>
						</div>
						<div class="flex items-center gap-3 shrink-0">
							<div class="text-right">
								<p class="font-semibold text-gray-800 text-sm">{formatPrice(item.base_price)}</p>
								<span
									class="text-xs px-2 py-0.5 rounded-full {item.is_active
										? 'bg-green-50 text-green-600'
										: 'bg-gray-100 text-gray-400'}"
								>
									{item.is_active ? 'Aktif' : 'Nonaktif'}
								</span>
							</div>
							<button
								onclick={() => {
									editingItem = item;
									showItemForm = false;
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
	{/if}
</div>
