<script lang="ts">
	import { page } from '$app/state';
	import { enhance } from '$app/forms';

	let { children, data } = $props();

	const navItems = [
		{ href: '/dashboard', label: 'Dashboard', icon: '🏠' },
		{ href: '/dashboard/pos', label: 'POS', icon: '🖥️' },
		{ href: '/dashboard/menu', label: 'Menu', icon: '🍽️' },
		{ href: '/dashboard/orders', label: 'Pesanan', icon: '📋' },
		{ href: '/dashboard/kds', label: 'Dapur', icon: '👨‍🍳' },
		{ href: '/dashboard/branches', label: 'Cabang', icon: '🏪' },
		{ href: '/dashboard/tables', label: 'Meja & QR', icon: '🪑' },
		{ href: '/dashboard/reports', label: 'Laporan', icon: '📊' },
		{ href: '/dashboard/settings', label: 'Pengaturan', icon: '⚙️' }
	];

	let sidebarOpen = $state(false);
</script>

<div class="flex h-screen bg-gray-100 overflow-hidden">
	<!-- Sidebar -->
	<aside
		class="w-64 flex-shrink-0 bg-white shadow-md flex flex-col transition-transform duration-200
      {sidebarOpen ? 'translate-x-0' : '-translate-x-full'}
      fixed inset-y-0 left-0 z-30 md:relative md:translate-x-0"
	>
		<div class="px-6 py-5 border-b border-gray-100">
			<span class="text-xl font-bold text-orange-600">RestoQRIS</span>
			{#if data.organization}
				<p class="text-xs text-gray-400 mt-0.5 truncate">{data.organization.name}</p>
			{:else}
				<p class="text-xs text-gray-400 mt-0.5">Multi-Cabang</p>
			{/if}
		</div>

		<nav class="flex-1 overflow-y-auto py-4 px-3">
			{#each navItems as item}
				<a
					href={item.href}
					class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium mb-1 transition-colors
            {page.url.pathname === item.href
						? 'bg-orange-50 text-orange-600'
						: 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'}"
				>
					<span class="text-lg">{item.icon}</span>
					{item.label}
				</a>
			{/each}
		</nav>

		<div class="px-4 py-4 border-t border-gray-100">
			<form method="POST" action="/logout" use:enhance>
				<button
					type="submit"
					class="flex items-center gap-2 text-sm text-gray-500 hover:text-red-500 transition-colors w-full"
				>
					<span>🚪</span> Keluar
				</button>
			</form>
		</div>
	</aside>

	<!-- Overlay for mobile -->
	{#if sidebarOpen}
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="fixed inset-0 z-20 bg-black/30 md:hidden"
			onclick={() => (sidebarOpen = false)}
		></div>
	{/if}

	<!-- Main content -->
	<div class="flex-1 flex flex-col overflow-hidden">
		<header class="bg-white shadow-sm px-4 py-3 flex items-center gap-4">
			<button
				class="md:hidden p-1.5 rounded-lg text-gray-500 hover:bg-gray-100"
				onclick={() => (sidebarOpen = !sidebarOpen)}
				aria-label="Toggle sidebar"
			>
				☰
			</button>
			<h1 class="text-sm font-semibold text-gray-700">
				Selamat datang, {data.user?.name ?? 'Admin'}
			</h1>
		</header>

		<main class="flex-1 overflow-y-auto p-6">
			{@render children()}
		</main>
	</div>
</div>
