<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { onDestroy } from 'svelte';
	import { API_BASE_URL, type KitchenTicket, type KitchenTicketStatus, type KitchenStation } from '$lib/api/client';
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let tickets = $state<KitchenTicket[]>(data.tickets);
	let branchId = $state<number | null>(data.selectedBranchId);
	let stationId = $state<number | null>(data.selectedStationId);
	let highlightedIds = $state<number[]>([]);
	let socket: WebSocket | null = null;
	let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

	$effect(() => {
		tickets = data.tickets;
		branchId = data.selectedBranchId;
		stationId = data.selectedStationId;
	});

	const stations = $derived(data.stations);
	const ticketColumns = $derived({
		queued: tickets.filter((ticket) => ticket.status === 'queued'),
		inProgress: tickets.filter((ticket) => ticket.status === 'in_progress'),
		done: tickets.filter((ticket) => ticket.status === 'done')
	});

	const statusMeta: Record<KitchenTicketStatus, { label: string; badge: string }> = {
		queued: { label: 'Baru', badge: 'bg-yellow-50 text-yellow-700' },
		in_progress: { label: 'Diproses', badge: 'bg-blue-50 text-blue-700' },
		done: { label: 'Siap', badge: 'bg-green-50 text-green-700' },
		cancelled: { label: 'Dibatalkan', badge: 'bg-red-50 text-red-600' }
	};

	function applyFilters() {
		const params = new URLSearchParams();
		if (branchId) params.set('branch_id', String(branchId));
		if (stationId) params.set('station_id', String(stationId));
		goto(`/dashboard/kds${params.toString() ? `?${params.toString()}` : ''}`);
	}

	function buildWebSocketURL(branch: number, token: string) {
		const url = new URL(API_BASE_URL);
		url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
		url.pathname = '/api/v1/kds/ws';
		url.searchParams.set('branch_id', String(branch));
		url.searchParams.set('token', token);
		return url.toString();
	}

	function connectWebSocket() {
		if (!branchId || !data.wsToken) return;
		const nextSocket = new WebSocket(buildWebSocketURL(branchId, data.wsToken));
		socket = nextSocket;

		nextSocket.onmessage = (event) => {
			const payload = JSON.parse(event.data) as { type?: string; ticket?: KitchenTicket };
			if (!payload.ticket) return;
			upsertTicket(payload.ticket);
			highlightTicket(payload.ticket.id);
			if (payload.type === 'ticket.created') playNotificationTone();
		};
		nextSocket.onclose = () => {
			if (socket === nextSocket) {
				socket = null;
			}
			if (branchId) {
				reconnectTimer = setTimeout(connectWebSocket, 2000);
			}
		};
	}

	function upsertTicket(ticket: KitchenTicket) {
		const matchesStation = !stationId || ticket.station_id === stationId;
		const nextTickets = tickets.filter((current) => current.id !== ticket.id);
		if (ticket.status !== 'cancelled' && matchesStation) {
			nextTickets.unshift(ticket);
		}
		tickets = nextTickets;
	}

	function highlightTicket(ticketId: number) {
		highlightedIds = Array.from(new Set([...highlightedIds, ticketId]));
		setTimeout(() => {
			highlightedIds = highlightedIds.filter((id) => id !== ticketId);
		}, 3500);
	}

	function playNotificationTone() {
		try {
			const AudioContextCtor = window.AudioContext || (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
			if (!AudioContextCtor) return;
			const context = new AudioContextCtor();
			const oscillator = context.createOscillator();
			const gain = context.createGain();
			oscillator.type = 'sine';
			oscillator.frequency.value = 880;
			gain.gain.value = 0.05;
			oscillator.connect(gain);
			gain.connect(context.destination);
			oscillator.start();
			setTimeout(() => {
				oscillator.stop();
				void context.close();
			}, 160);
		} catch {
			// noop
		}
	}

	$effect(() => {
		if (!branchId || !data.wsToken) return;
		connectWebSocket();
		return () => {
			if (reconnectTimer) clearTimeout(reconnectTimer);
			reconnectTimer = null;
			socket?.close();
			socket = null;
		};
	});

	onDestroy(() => {
		if (reconnectTimer) clearTimeout(reconnectTimer);
		socket?.close();
	});

	function formatTime(value?: string) {
		if (!value) return '—';
		return new Date(value).toLocaleTimeString('id-ID', {
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function stationLabel(ticket: KitchenTicket) {
		return ticket.station_name ?? 'Belum ditentukan';
	}

	function isHighlighted(ticketId: number) {
		return highlightedIds.includes(ticketId);
	}

	const stationOptions = $derived(stations as KitchenStation[]);
</script>

<svelte:head>
	<title>KDS — RestoQRIS</title>
</svelte:head>

<div class="max-w-7xl mx-auto space-y-6">
	<div class="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
		<div>
			<h2 class="text-2xl font-bold text-gray-800">Kitchen Display System</h2>
			<p class="text-sm text-gray-500">Pantau antrian dapur secara real-time per cabang.</p>
		</div>
		<div class="text-xs text-gray-400">
			{#if socket}
				<span class="inline-flex items-center gap-1 rounded-full bg-green-50 px-2.5 py-1 text-green-600">● Realtime aktif</span>
			{:else}
				<span class="inline-flex items-center gap-1 rounded-full bg-yellow-50 px-2.5 py-1 text-yellow-700">● Menyambung...</span>
			{/if}
		</div>
	</div>

	{#if form && !form.success && 'error' in form && form.error}
		<div class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{form.error}</div>
	{/if}
	{#if form?.success}
		<div class="rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700">Perubahan KDS berhasil disimpan.</div>
	{/if}

	<div class="grid gap-4 lg:grid-cols-[2fr,1fr]">
		<div class="rounded-xl bg-white p-4 shadow-sm">
			<div class="grid gap-3 md:grid-cols-[1fr,1fr,auto] md:items-end">
				<div>
					<label for="branch-filter" class="mb-1 block text-xs font-medium text-gray-500">Cabang</label>
					<select id="branch-filter" bind:value={branchId} class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400">
						{#if data.branches.length === 0}
							<option value={null}>Belum ada cabang</option>
						{:else}
							{#each data.branches as branch}
								<option value={branch.id}>{branch.name}</option>
							{/each}
						{/if}
					</select>
				</div>
				<div>
					<label for="station-filter" class="mb-1 block text-xs font-medium text-gray-500">Station</label>
					<select id="station-filter" bind:value={stationId} class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400">
						<option value={null}>Semua Station</option>
						{#each stationOptions as station}
							<option value={station.id}>{station.name}</option>
						{/each}
					</select>
				</div>
				<button onclick={applyFilters} class="rounded-lg bg-orange-500 px-4 py-2 text-sm font-semibold text-white hover:bg-orange-600">Terapkan</button>
			</div>
		</div>

		<div class="rounded-xl bg-white p-4 shadow-sm">
			<h3 class="mb-3 text-sm font-semibold text-gray-800">Tambah Station</h3>
			<form method="POST" action="?/createStation" use:enhance={() => {
				return async ({ update }) => {
					await update();
				};
			}} class="space-y-3">
				<input type="hidden" name="branch_id" value={branchId ?? ''} />
				<div>
					<label for="station-name" class="mb-1 block text-xs font-medium text-gray-500">Nama Station</label>
					<input id="station-name" name="name" type="text" placeholder="Contoh: Grill" class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400" />
				</div>
				<button type="submit" class="w-full rounded-lg bg-gray-900 px-4 py-2 text-sm font-semibold text-white hover:bg-gray-800 disabled:cursor-not-allowed disabled:bg-gray-300" disabled={!branchId}>Buat Station</button>
			</form>
		</div>
	</div>

	{#if !branchId}
		<div class="rounded-xl bg-white p-12 text-center shadow-sm">
			<span class="mb-4 block text-5xl">👨‍🍳</span>
			<p class="text-gray-500">Belum ada cabang aktif untuk KDS.</p>
		</div>
	{:else}
		<div class="grid gap-4 xl:grid-cols-3">
			<div class="rounded-xl bg-white p-4 shadow-sm">
				<div class="mb-4 flex items-center justify-between">
					<h3 class="font-semibold text-gray-800">Baru Masuk</h3>
					<span class="rounded-full bg-yellow-50 px-2 py-0.5 text-xs font-medium text-yellow-700">{ticketColumns.queued.length}</span>
				</div>
				<div class="space-y-3">
					{#if ticketColumns.queued.length === 0}
						<p class="rounded-lg border border-dashed border-gray-200 px-4 py-8 text-center text-sm text-gray-400">Belum ada ticket baru.</p>
					{:else}
						{#each ticketColumns.queued as ticket (ticket.id)}
							<div class={`rounded-xl border p-4 transition ${isHighlighted(ticket.id) ? 'border-orange-300 bg-orange-50 shadow-md' : 'border-gray-100 bg-gray-50'}`}>
								<div class="mb-2 flex items-start justify-between gap-3">
									<div>
										<p class="font-semibold text-gray-800">{ticket.order_code}</p>
										<p class="text-xs text-gray-400">{stationLabel(ticket)}{#if ticket.table_id} · Meja #{ticket.table_id}{/if}</p>
									</div>
									<span class={`rounded-full px-2 py-0.5 text-xs font-medium ${statusMeta[ticket.status].badge}`}>{statusMeta[ticket.status].label}</span>
								</div>
								<p class="text-sm font-medium text-gray-700">{ticket.item_name}</p>
								<p class="mt-1 text-xs text-gray-500">Qty {ticket.quantity} · Masuk {formatTime(ticket.created_at)}</p>
								{#if ticket.notes}
									<p class="mt-2 rounded-lg bg-white px-3 py-2 text-xs italic text-gray-500">{ticket.notes}</p>
								{/if}
								<form method="POST" action="?/updateTicketStatus" use:enhance={() => {
									return async ({ update }) => {
										await update();
									};
								}} class="mt-3">
									<input type="hidden" name="branch_id" value={branchId} />
									<input type="hidden" name="ticket_id" value={ticket.id} />
									<input type="hidden" name="status" value="in_progress" />
									<button type="submit" class="w-full rounded-lg bg-orange-500 px-3 py-2 text-xs font-semibold text-white hover:bg-orange-600">Mulai Proses</button>
								</form>
							</div>
						{/each}
					{/if}
				</div>
			</div>

			<div class="rounded-xl bg-white p-4 shadow-sm">
				<div class="mb-4 flex items-center justify-between">
					<h3 class="font-semibold text-gray-800">Sedang Diproses</h3>
					<span class="rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700">{ticketColumns.inProgress.length}</span>
				</div>
				<div class="space-y-3">
					{#if ticketColumns.inProgress.length === 0}
						<p class="rounded-lg border border-dashed border-gray-200 px-4 py-8 text-center text-sm text-gray-400">Belum ada ticket diproses.</p>
					{:else}
						{#each ticketColumns.inProgress as ticket (ticket.id)}
							<div class={`rounded-xl border p-4 transition ${isHighlighted(ticket.id) ? 'border-orange-300 bg-orange-50 shadow-md' : 'border-gray-100 bg-gray-50'}`}>
								<div class="mb-2 flex items-start justify-between gap-3">
									<div>
										<p class="font-semibold text-gray-800">{ticket.order_code}</p>
										<p class="text-xs text-gray-400">{stationLabel(ticket)} · Mulai {formatTime(ticket.started_at)}</p>
									</div>
									<span class={`rounded-full px-2 py-0.5 text-xs font-medium ${statusMeta[ticket.status].badge}`}>{statusMeta[ticket.status].label}</span>
								</div>
								<p class="text-sm font-medium text-gray-700">{ticket.item_name}</p>
								<p class="mt-1 text-xs text-gray-500">Qty {ticket.quantity}</p>
								<form method="POST" action="?/updateTicketStatus" use:enhance={() => {
									return async ({ update }) => {
										await update();
									};
								}} class="mt-3">
									<input type="hidden" name="branch_id" value={branchId} />
									<input type="hidden" name="ticket_id" value={ticket.id} />
									<input type="hidden" name="status" value="done" />
									<button type="submit" class="w-full rounded-lg bg-blue-600 px-3 py-2 text-xs font-semibold text-white hover:bg-blue-700">Tandai Siap</button>
								</form>
							</div>
						{/each}
					{/if}
				</div>
			</div>

			<div class="rounded-xl bg-white p-4 shadow-sm">
				<div class="mb-4 flex items-center justify-between">
					<h3 class="font-semibold text-gray-800">Selesai</h3>
					<span class="rounded-full bg-green-50 px-2 py-0.5 text-xs font-medium text-green-700">{ticketColumns.done.length}</span>
				</div>
				<div class="space-y-3">
					{#if ticketColumns.done.length === 0}
						<p class="rounded-lg border border-dashed border-gray-200 px-4 py-8 text-center text-sm text-gray-400">Belum ada ticket selesai.</p>
					{:else}
						{#each ticketColumns.done as ticket (ticket.id)}
							<div class={`rounded-xl border p-4 transition ${isHighlighted(ticket.id) ? 'border-orange-300 bg-orange-50 shadow-md' : 'border-gray-100 bg-gray-50'}`}>
								<div class="mb-2 flex items-start justify-between gap-3">
									<div>
										<p class="font-semibold text-gray-800">{ticket.order_code}</p>
										<p class="text-xs text-gray-400">{stationLabel(ticket)} · Selesai {formatTime(ticket.completed_at)}</p>
									</div>
									<span class={`rounded-full px-2 py-0.5 text-xs font-medium ${statusMeta[ticket.status].badge}`}>{statusMeta[ticket.status].label}</span>
								</div>
								<p class="text-sm font-medium text-gray-700">{ticket.item_name}</p>
								<p class="mt-1 text-xs text-gray-500">Qty {ticket.quantity}</p>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>
