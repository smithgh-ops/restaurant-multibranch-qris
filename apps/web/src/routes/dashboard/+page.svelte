<script lang="ts">
	const modules = [
		{
			title: 'Point of Sale (POS)',
			icon: '🖥️',
			description: 'Proses transaksi meja & takeaway dengan pembayaran QRIS.',
			status: 'Segera Hadir'
		},
		{
			title: 'Manajemen Menu',
			icon: '🍽️',
			description: 'Kelola menu, kategori, varian, dan harga per-cabang.',
			status: 'Segera Hadir'
		},
		{
			title: 'Pesanan',
			icon: '📋',
			description: 'Monitor semua pesanan aktif, riwayat, dan statusnya.',
			status: 'Segera Hadir'
		},
		{
			title: 'Dapur (KDS)',
			icon: '👨‍🍳',
			description: 'Kitchen Display System untuk manajemen antrian masak.',
			status: 'Segera Hadir'
		},
		{
			title: 'Manajemen Cabang',
			icon: '🏪',
			description: 'Kelola semua cabang restoran beserta pengaturannya.',
			status: 'Segera Hadir'
		},
		{
			title: 'Laporan & Analitik',
			icon: '📊',
			description: 'Laporan penjualan, performa cabang, dan produk terlaris.',
			status: 'Segera Hadir'
		},
		{
			title: 'Pembayaran QRIS',
			icon: '💳',
			description: 'Integrasi payment gateway QRIS dengan webhook verification.',
			status: 'Segera Hadir'
		},
		{
			title: 'Pengaturan',
			icon: '⚙️',
			description: 'Konfigurasi sistem, pengguna, role, dan preferensi.',
			status: 'Segera Hadir'
		}
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

	<!-- Stats placeholder -->
	<div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
		{#each [['Cabang', '—'], ['Pesanan Hari Ini', '—'], ['Pendapatan', '—'], ['Meja Aktif', '—']] as [label, value]}
			<div class="bg-white rounded-xl shadow-sm p-5">
				<p class="text-xs text-gray-400 font-medium uppercase tracking-wider">{label}</p>
				<p class="mt-2 text-2xl font-bold text-gray-700">{value}</p>
			</div>
		{/each}
	</div>

	<!-- Module grid -->
	<h3 class="text-lg font-semibold text-gray-700 mb-4">Modul MVP</h3>
	<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
		{#each modules as mod}
			<div class="bg-white rounded-xl shadow-sm p-5 hover:shadow-md transition-shadow">
				<div class="text-3xl mb-3">{mod.icon}</div>
				<h4 class="font-semibold text-gray-800 text-sm">{mod.title}</h4>
				<p class="text-xs text-gray-500 mt-1 leading-relaxed">{mod.description}</p>
				<span
					class="inline-block mt-3 text-xs font-medium px-2 py-0.5 rounded-full bg-orange-50 text-orange-500"
				>
					{mod.status}
				</span>
			</div>
		{/each}
	</div>
</div>
