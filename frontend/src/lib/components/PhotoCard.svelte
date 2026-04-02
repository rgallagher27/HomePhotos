<script lang="ts">
	import type { PhotoListItem } from '$lib/api/gen/types.gen';
	import { thumbUrl } from '$lib/image';

	let { photo, onclick }: { photo: PhotoListItem; onclick?: () => void } = $props();

	function formatDate(dateStr: string | null | undefined): string {
		if (!dateStr) return '';
		return new Date(dateStr).toLocaleDateString(undefined, {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}
</script>

<button
	type="button"
	class="group relative aspect-square overflow-hidden rounded bg-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500"
	{onclick}
>
	<img
		src={thumbUrl(photo.id)}
		alt={photo.file_name}
		loading="lazy"
		class="h-full w-full object-cover transition-transform duration-200 group-hover:scale-105"
	/>
	<div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/60 to-transparent p-2 opacity-0 transition-opacity group-hover:opacity-100">
		<p class="truncate text-xs text-white">{photo.file_name}</p>
		{#if photo.captured_at}
			<p class="text-xs text-white/70">{formatDate(photo.captured_at)}</p>
		{/if}
	</div>
</button>
