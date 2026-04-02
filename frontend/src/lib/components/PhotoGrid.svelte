<script lang="ts">
	import type { PhotoListItem } from '$lib/api/gen/types.gen';
	import PhotoCard from './PhotoCard.svelte';

	let {
		photos,
		hasMore,
		loading,
		onLoadMore,
		onPhotoClick
	}: {
		photos: PhotoListItem[];
		hasMore: boolean;
		loading: boolean;
		onLoadMore: () => void;
		onPhotoClick: (photo: PhotoListItem) => void;
	} = $props();

	let sentinel: HTMLDivElement | undefined = $state();

	$effect(() => {
		if (!sentinel) return;

		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting && hasMore && !loading) {
					onLoadMore();
				}
			},
			{ rootMargin: '200px' }
		);

		observer.observe(sentinel);
		return () => observer.disconnect();
	});

	type DateGroup = { date: string; label: string; photos: PhotoListItem[] };

	const groups: DateGroup[] = $derived.by(() => {
		const map = new Map<string, PhotoListItem[]>();
		for (const photo of photos) {
			const dateKey = photo.captured_at
				? new Date(photo.captured_at).toISOString().slice(0, 10)
				: 'unknown';
			const existing = map.get(dateKey);
			if (existing) {
				existing.push(photo);
			} else {
				map.set(dateKey, [photo]);
			}
		}

		return Array.from(map.entries()).map(([dateKey, items]) => ({
			date: dateKey,
			label:
				dateKey === 'unknown'
					? 'No date'
					: new Date(dateKey + 'T00:00:00').toLocaleDateString(undefined, {
							weekday: 'long',
							month: 'long',
							day: 'numeric',
							year: 'numeric'
						}),
			photos: items
		}));
	});
</script>

<div class="space-y-6">
	{#each groups as group (group.date)}
		<section>
			<h3 class="mb-2 text-sm font-medium text-gray-500">{group.label}</h3>
			<div class="grid grid-cols-[repeat(auto-fill,minmax(180px,1fr))] gap-1">
				{#each group.photos as photo (photo.id)}
					<PhotoCard {photo} onclick={() => onPhotoClick(photo)} />
				{/each}
			</div>
		</section>
	{/each}

	{#if loading}
		<div class="flex justify-center py-4">
			<div class="text-sm text-gray-400">Loading...</div>
		</div>
	{/if}

	{#if !loading && photos.length === 0}
		<div class="flex justify-center py-12">
			<p class="text-gray-400">No photos found</p>
		</div>
	{/if}

	<div bind:this={sentinel} class="h-1"></div>
</div>
