<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib';

	void api;

	let email = $state('');
	let password = $state('');
	let loading = $state(false);
	let errorMsg = $state('');

	async function handleLogin(e: SubmitEvent) {
		e.preventDefault();
		loading = true;
		errorMsg = '';
		// TODO: replace with real auth endpoint
		await new Promise((r) => setTimeout(r, 500));
		loading = false;
		if (email && password) {
			goto('/dashboard');
		} else {
			errorMsg = 'Email dan password wajib diisi.';
		}
	}
</script>

<svelte:head>
	<title>Login — RestoQRIS</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-gray-50 px-4">
	<div class="w-full max-w-md">
		<div class="text-center mb-8">
			<h1 class="text-3xl font-bold text-orange-600">RestoQRIS</h1>
			<p class="mt-2 text-gray-500 text-sm">Manajemen Restoran Multi-Cabang</p>
		</div>

		<div class="bg-white rounded-2xl shadow-md p-8">
			<h2 class="text-xl font-semibold text-gray-800 mb-6">Masuk ke Akun Anda</h2>

			{#if errorMsg}
				<div class="mb-4 rounded-lg bg-red-50 border border-red-200 text-red-700 px-4 py-3 text-sm">
					{errorMsg}
				</div>
			{/if}

			<form onsubmit={handleLogin} class="space-y-4">
				<div>
					<label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
					<input
						id="email"
						type="email"
						bind:value={email}
						required
						placeholder="nama@restoran.com"
						class="w-full rounded-lg border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent"
					/>
				</div>

				<div>
					<label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label
					>
					<input
						id="password"
						type="password"
						bind:value={password}
						required
						placeholder="••••••••"
						class="w-full rounded-lg border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent"
					/>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="w-full rounded-lg bg-orange-500 hover:bg-orange-600 disabled:opacity-60 text-white font-semibold py-2.5 text-sm transition-colors"
				>
					{loading ? 'Memproses...' : 'Masuk'}
				</button>
			</form>
		</div>
	</div>
</div>
