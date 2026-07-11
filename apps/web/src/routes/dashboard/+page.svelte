<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const modules = [
		{
			title: 'Point of Sale (POS)',
			icon: '🖥️',
			description: 'Proses transaksi meja & takeaway dengan pembayaran QRIS.',
			href: '/dashboard/pos',
			status: 'Segera Hadir'
		},
		{
			title: 'Manajemen Menu',
			icon: '🍽️',
			description: 'Kelola menu, kategori, varian, dan harga per-cabang.',
			href: '/dashboard/menu',
			status: 'Aktif'
		},
		{
			title: 'Pesanan',
			icon: '📋',
			description: 'Monitor semua pesanan aktif, riwayat, dan statusnya.',
			href: '/dashboard/orders',
			status: 'Segera Hadir'
		},
		{
			title: 'Dapur (KDS)',
			icon: '👨‍🍳',
			description: 'Kitchen Display System untuk manajemen antrian masak.',
			href: '/dashboard/kds',
			status: 'Aktif'
		},
		{
			title: 'Manajemen Cabang',
			icon: '🏪',
			description: 'Kelola semua cabang restoran beserta pengaturannya.',
			href: '/dashboard/branches',
			status: 'Aktif'
		},
		{
			title: 'Laporan & Analitik',
			icon: '📊',
			description: 'Laporan penjualan, performa cabang, dan produk terlaris.',
			href: '/dashboard/reports',
			status: 'Aktif'
		},
		{
			title: 'Pembayaran QRIS',
			icon: '💳',
			description: 'Integrasi payment gateway QRIS dengan webhook verification.',
			href: '#',
			status: 'Segera Hadir'
		},
		{
			title: 'Pengaturan',
			icon: '⚙️',
			description: 'Konfigurasi sistem, pengguna, role, dan preferensi.',
			href: '/dashboard/settings',
			status: 'Segera Hadir'
		}
	];

	const stats = [
		{ label: 'Cabang', value: data.branches.length, icon: '🏪' },
		{ label: 'Kategori Menu', value: data.categories.length, icon: '🍽️' },
		{ label: 'Item Menu', value: data.menuItems.length, icon: '🍜' },
		{ label: 'Pesanan Hari Ini', value: '—', icon: '📋' }
	];
</script>

<svelte:head>
	<title>Dashboard — RestoQRIS</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
	<div class="mb-8">
		<h2 class="text-2xl font-bold text-gray-800">RestoQRIS — Manajemen Restoran Multi-Cabang</h2>
		<p class="mt-1 text-gray-500 text-sm">
			Platform terpusat untuk operasional restoran dengan dukungan pembayaran QRIS.
		</p>
	</div>

	<!-- Stats -->
	<div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
		{#each stats as stat}
			<div class="bg-white rounded-xl shadow-sm p-5">
				<div class="flex items-center gap-2 mb-2">
					<span class="text-2xl">{stat.icon}</span>
					<p class="text-xs text-gray-400 font-medium uppercase tracking-wider">{stat.label}</p>
				</div>
				<p class="text-2xl font-bold text-gray-700">{stat.value}</p>
			</div>
		{/each}
	</div>

	<!-- Branches quick view -->
	{#if data.branches.length > 0}
		<div class="mb-8">
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-lg font-semibold text-gray-700">Cabang Aktif</h3>
				<a href="/dashboard/branches" class="text-sm text-orange-500 hover:text-orange-600"
					>Lihat semua →</a
				>
			</div>
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
				{#each data.branches.slice(0, 3) as branch}
					<div class="bg-white rounded-xl shadow-sm p-4 flex items-start gap-3">
						<span class="text-2xl">🏪</span>
						<div>
							<p class="font-semibold text-gray-800 text-sm">{branch.name}</p>
							{#if branch.address}
								<p class="text-xs text-gray-500 mt-0.5 line-clamp-1">{branch.address}</p>
							{/if}
							<span
								class="inline-block mt-1 text-xs px-2 py-0.5 rounded-full {branch.is_active
									? 'bg-green-50 text-green-600'
									: 'bg-gray-100 text-gray-400'}"
							>
								{branch.is_active ? 'Aktif' : 'Nonaktif'}
							</span>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Module grid -->
	<h3 class="text-lg font-semibold text-gray-700 mb-4">Modul</h3>
	<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
		{#each modules as mod}
			<a
				href={mod.href}
				class="bg-white rounded-xl shadow-sm p-5 hover:shadow-md transition-shadow block"
				class:pointer-events-none={mod.status !== 'Aktif' && mod.href === '#'}
			>
				<div class="text-3xl mb-3">{mod.icon}</div>
				<h4 class="font-semibold text-gray-800 text-sm">{mod.title}</h4>
				<p class="text-xs text-gray-500 mt-1 leading-relaxed">{mod.description}</p>
				<span
					class="inline-block mt-3 text-xs font-medium px-2 py-0.5 rounded-full {mod.status ===
					'Aktif'
						? 'bg-green-50 text-green-600'
						: 'bg-orange-50 text-orange-500'}"
				>
					{mod.status}
				</span>
			</a>
		{/each}
	</div>
</div>
