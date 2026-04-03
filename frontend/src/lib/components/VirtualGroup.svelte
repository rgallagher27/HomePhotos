<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();
	let wrapper: HTMLDivElement | undefined = $state();
	let isVisible = $state(true);
	let measuredHeight = $state(0);

	$effect(() => {
		if (!wrapper) return;
		const observer = new IntersectionObserver(
			([entry]) => {
				if (entry.isIntersecting) {
					isVisible = true;
				} else if (measuredHeight > 0 || wrapper!.offsetHeight > 0) {
					measuredHeight = wrapper!.offsetHeight;
					isVisible = false;
				}
			},
			{ rootMargin: '600px' }
		);
		observer.observe(wrapper);
		return () => observer.disconnect();
	});
</script>

<div bind:this={wrapper}>
	{#if isVisible}
		{@render children()}
	{:else}
		<div style="height: {measuredHeight}px"></div>
	{/if}
</div>
