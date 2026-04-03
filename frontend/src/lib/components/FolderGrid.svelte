<script lang="ts">
	let {
		folders,
		onFolderClick
	}: {
		folders: string[];
		onFolderClick: (name: string) => void;
	} = $props();
</script>

{#if folders.length > 0}
	<div class="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-2">
		{#each folders as folder (folder)}
			<button
				type="button"
				onclick={() => onFolderClick(folder)}
				class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2.5 text-left text-sm text-gray-700 hover:border-gray-300 hover:bg-gray-50 transition-colors"
			>
				<svg class="h-5 w-5 shrink-0 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
				</svg>
				<span class="truncate">{folder}</span>
			</button>
		{/each}
	</div>
{/if}
