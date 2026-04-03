<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { ScannerStatusResponse } from '$lib/api/gen/types.gen';
	import { getScannerStatus, postScannerRun } from '$lib/api/gen/sdk.gen';

	let status: ScannerStatusResponse | null = $state(null);
	let running = $state(false);
	let pollTimer: ReturnType<typeof setInterval> | null = null;
	let error = $state('');

	async function fetchStatus() {
		const res = await getScannerStatus();
		if (res.error) {
			error = 'Failed to fetch scanner status';
			return;
		}
		status = res.data as unknown as ScannerStatusResponse;
		running = status.status === 'scanning';
	}

	async function startScan() {
		error = '';
		const res = await postScannerRun();
		if (res.error) {
			const errData = res.error as unknown as { error?: { message?: string } };
			error = errData?.error?.message ?? 'Scan already in progress';
			return;
		}
		running = true;
		startPolling();
	}

	function startPolling() {
		stopPolling();
		pollTimer = setInterval(async () => {
			await fetchStatus();
			if (!running) stopPolling();
		}, 2000);
	}

	function stopPolling() {
		if (pollTimer) {
			clearInterval(pollTimer);
			pollTimer = null;
		}
	}

	onMount(async () => {
		await fetchStatus();
		if (running) startPolling();
	});

	onDestroy(() => stopPolling());

	const progress = $derived.by(() => {
		if (!status || status.total_files === 0) return 0;
		return Math.round((status.processed / status.total_files) * 100);
	});
</script>

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<h3 class="text-lg font-medium text-gray-900">Scanner</h3>
		<button
			type="button"
			onclick={startScan}
			disabled={running}
			class="rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
		>
			{running ? 'Scanning...' : 'Run Scan'}
		</button>
	</div>

	{#if error}
		<div class="rounded bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}

	{#if status}
		<div class="rounded border border-gray-200 p-4 space-y-3">
			<div class="flex items-center gap-2">
				<span class="text-sm text-gray-500">Status:</span>
				<span
					class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium
					{running ? 'bg-blue-100 text-blue-700' : 'bg-gray-100 text-gray-600'}"
				>
					{status.status}
				</span>
			</div>

			{#if running && status.total_files > 0}
				<div>
					<div class="flex justify-between text-sm text-gray-500 mb-1">
						<span>{status.processed} / {status.total_files}</span>
						<span>{progress}%</span>
					</div>
					<div class="h-2 rounded-full bg-gray-200 overflow-hidden">
						<div
							class="h-full rounded-full bg-blue-600 transition-all duration-300"
							style="width: {progress}%"
						></div>
					</div>
				</div>
			{/if}

			{#if status.errors > 0}
				<p class="text-sm text-red-600">{status.errors} error(s)</p>
			{/if}

			{#if status.started_at}
				<p class="text-xs text-gray-400">
					Started: {new Date(status.started_at).toLocaleString()}
				</p>
			{/if}

			{#if !running && status.started_at}
				<div class="rounded bg-gray-50 px-4 py-3 text-sm text-gray-700 space-y-1">
					<p class="font-medium">Last scan results</p>
					<div class="grid grid-cols-2 gap-x-4 gap-y-1 text-xs text-gray-600">
						<span>{status.added} new</span>
						<span>{status.updated} updated</span>
						<span>{status.unchanged} unchanged</span>
						<span>{status.deleted} removed</span>
						<span>{status.errors} error(s)</span>
					</div>
					{#if status.skipped > 0}
						<p class="text-xs text-amber-600">{status.skipped} file(s) skipped (zero-byte)</p>
					{/if}
				</div>
			{/if}
		</div>
	{/if}
</div>
