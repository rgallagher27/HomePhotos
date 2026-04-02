<script lang="ts">
	let {
		name,
		color,
		selected = false,
		removable = false,
		onclick,
		onremove
	}: {
		name: string;
		color?: string | null;
		selected?: boolean;
		removable?: boolean;
		onclick?: () => void;
		onremove?: () => void;
	} = $props();

	const baseClass = $derived(
		selected
			? color
				? 'text-white'
				: 'bg-blue-100 text-blue-800'
			: color
				? 'text-white opacity-60 hover:opacity-100'
				: 'bg-gray-100 text-gray-700 hover:bg-gray-200'
	);
</script>

<button
	type="button"
	{onclick}
	class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium transition-opacity {baseClass}"
	style={color ? `background-color: ${color}` : ''}
>
	{name}
	{#if removable && onremove}
		<span
			role="button"
			tabindex="0"
			onclick={(e) => { e.stopPropagation(); onremove?.(); }}
			onkeydown={(e) => { if (e.key === 'Enter') { e.stopPropagation(); onremove?.(); } }}
			class="ml-0.5 hover:opacity-70"
			aria-label="Remove {name}"
		>×</span>
	{/if}
</button>
