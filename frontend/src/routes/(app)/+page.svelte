<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import type { PhotoListItem, PhotoListResponse } from '$lib/api/gen/types.gen';
	import { getPhotos } from '$lib/api/gen/sdk.gen';
	import PhotoGrid from '$lib/components/PhotoGrid.svelte';

	let photos: PhotoListItem[] = $state([]);
	let cursor: string | null = $state(null);
	let hasMore = $state(true);
	let loading = $state(false);

	async function loadMore() {
		if (loading || !hasMore) return;
		loading = true;

		const query: Record<string, unknown> = { limit: 50, sort: 'captured_at', order: 'desc' };
		if (cursor) query.cursor = cursor;

		const res = await getPhotos({ query });

		if (res.error) {
			loading = false;
			return;
		}

		const data = res.data as unknown as PhotoListResponse;
		photos = [...photos, ...data.data];
		cursor = data.next_cursor ?? null;
		hasMore = data.has_more;
		loading = false;
	}

	function handlePhotoClick(photo: PhotoListItem) {
		goto(`/photos/${photo.id}`);
	}

	onMount(() => {
		loadMore();
	});
</script>

<div class="p-4">
	<PhotoGrid {photos} {hasMore} {loading} onLoadMore={loadMore} onPhotoClick={handlePhotoClick} />
</div>
