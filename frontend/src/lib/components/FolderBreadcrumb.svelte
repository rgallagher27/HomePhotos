<script lang="ts">
	let {
		path,
		onNavigate
	}: {
		path: string;
		onNavigate: (path: string) => void;
	} = $props();

	const segments: { label: string; path: string }[] = $derived.by(() => {
		if (!path) return [];
		const parts = path.split('/');
		return parts.map((part, i) => ({
			label: part,
			path: parts.slice(0, i + 1).join('/')
		}));
	});
</script>

<nav class="flex items-center gap-1 text-sm text-gray-500">
	<button
		type="button"
		onclick={() => onNavigate('')}
		class="hover:text-gray-900 {path ? '' : 'font-medium text-gray-900'}"
	>
		All folders
	</button>
	{#each segments as segment (segment.path)}
		<span class="text-gray-300">/</span>
		<button
			type="button"
			onclick={() => onNavigate(segment.path)}
			class="hover:text-gray-900 {segment.path === path ? 'font-medium text-gray-900' : ''}"
		>
			{segment.label}
		</button>
	{/each}
</nav>
