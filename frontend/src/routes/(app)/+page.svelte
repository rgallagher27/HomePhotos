<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import type { PhotoListItem, PhotoListResponse } from '$lib/api/gen/types.gen';
	import { getPhotos } from '$lib/api/gen/sdk.gen';
	import PhotoGrid from '$lib/components/PhotoGrid.svelte';
	import TagSidebar from '$lib/components/TagSidebar.svelte';
	import TagChip from '$lib/components/TagChip.svelte';
	import { setPhotoNav } from '$lib/photoNav.svelte';

	let photos: PhotoListItem[] = $state([]);
	let cursor: string | null = $state(null);
	let hasMore = $state(true);
	let loading = $state(false);

	let error = $state('');
	let selectedTagIds: number[] = $state([]);
	let tagMode: 'and' | 'or' = $state('or');
	let sidebarOpen = $state(false);

	async function loadMore() {
		if (loading || !hasMore) return;
		loading = true;

		const query: Record<string, unknown> = { limit: 50, sort: 'captured_at', order: 'desc' };
		if (cursor) query.cursor = cursor;
		if (selectedTagIds.length > 0) {
			query.tags = selectedTagIds.join(',');
			query.tag_mode = tagMode;
		}

		const res = await getPhotos({ query });

		if (res.error) {
			error = 'Failed to load photos';
			loading = false;
			return;
		}
		error = '';

		const data = res.data as unknown as PhotoListResponse;
		photos = [...photos, ...data.data];
		cursor = data.next_cursor ?? null;
		hasMore = data.has_more;
		loading = false;
	}

	function resetAndLoad() {
		photos = [];
		cursor = null;
		hasMore = true;
		loadMore();
	}

	function toggleTag(id: number) {
		if (selectedTagIds.includes(id)) {
			selectedTagIds = selectedTagIds.filter((t) => t !== id);
		} else {
			selectedTagIds = [...selectedTagIds, id];
		}
		resetAndLoad();
	}

	function toggleMode() {
		tagMode = tagMode === 'or' ? 'and' : 'or';
		if (selectedTagIds.length > 0) resetAndLoad();
	}

	function clearTags() {
		selectedTagIds = [];
		resetAndLoad();
	}

	function handlePhotoClick(photo: PhotoListItem) {
		setPhotoNav(photos.map((p) => p.id));
		goto(`/photos/${photo.id}`);
	}

	onMount(() => {
		loadMore();
	});

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && sidebarOpen) sidebarOpen = false;
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="flex h-[calc(100vh-49px)]">
	<!-- Mobile sidebar toggle -->
	<button
		type="button"
		onclick={() => (sidebarOpen = !sidebarOpen)}
		class="fixed bottom-4 left-4 z-30 rounded-full bg-white p-3 shadow-lg border border-gray-200 lg:hidden"
		aria-label="Toggle tag filters"
	>
		<svg class="h-5 w-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
		</svg>
		{#if selectedTagIds.length > 0}
			<span class="absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full bg-blue-600 text-[10px] text-white">
				{selectedTagIds.length}
			</span>
		{/if}
	</button>

	<!-- Mobile backdrop -->
	{#if sidebarOpen}
		<button
			type="button"
			class="fixed inset-0 z-20 bg-black/30 lg:hidden cursor-default"
			onclick={() => (sidebarOpen = false)}
			aria-label="Close sidebar"
		></button>
	{/if}

	<!-- Sidebar -->
	<aside
		class="fixed inset-y-0 left-0 z-20 w-64 transform overflow-y-auto border-r border-gray-200 bg-white p-4 pt-16 transition-transform duration-200 lg:static lg:z-auto lg:translate-x-0 lg:pt-4
		{sidebarOpen ? 'translate-x-0' : '-translate-x-full'}"
	>
		<TagSidebar
			{selectedTagIds}
			{tagMode}
			onToggleTag={toggleTag}
			onToggleMode={toggleMode}
			onClear={clearTags}
		/>
	</aside>

	<!-- Main content -->
	<div class="flex-1 overflow-y-auto p-4">
		{#if error}
			<div class="mb-4 rounded bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
		{/if}

		{#if selectedTagIds.length > 0}
			<div class="mb-4 flex flex-wrap items-center gap-2">
				<span class="text-xs text-gray-500">Filtering:</span>
				{#each selectedTagIds as id (id)}
					<TagChip name={`Tag #${id}`} selected removable onremove={() => toggleTag(id)} />
				{/each}
				<button type="button" onclick={clearTags} class="text-xs text-gray-400 hover:text-gray-600">
					Clear all
				</button>
			</div>
		{/if}

		<PhotoGrid {photos} {hasMore} {loading} onLoadMore={loadMore} onPhotoClick={handlePhotoClick} />
	</div>
</div>
