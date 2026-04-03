<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import type { PhotoListItem, PhotoListResponse, FolderListResponse } from '$lib/api/gen/types.gen';
	import { getPhotos, getFolders } from '$lib/api/gen/sdk.gen';
	import PhotoGrid from '$lib/components/PhotoGrid.svelte';
	import TagSidebar from '$lib/components/TagSidebar.svelte';
	import TagChip from '$lib/components/TagChip.svelte';
	import FolderBreadcrumb from '$lib/components/FolderBreadcrumb.svelte';
	import FolderGrid from '$lib/components/FolderGrid.svelte';
	import { setPhotoNav } from '$lib/photoNav.svelte';
	import * as Sheet from '$lib/components/ui/sheet/index.js';

	let photos: PhotoListItem[] = $state([]);
	let cursor: string | null = $state(null);
	let hasMore = $state(true);
	let loading = $state(false);

	let error = $state('');
	let selectedTagIds: number[] = $state([]);
	let tagMode: 'and' | 'or' = $state('or');
	let sidebarOpen = $state(false);
	let groupBy = $state<'date' | 'folder'>('date');

	// Folder browsing state
	let currentFolder = $state('');
	let subfolders: string[] = $state([]);
	let folderPhotoCount = $state(0);

	const isFolderMode = $derived(groupBy === 'folder');

	async function loadFolders() {
		const res = await getFolders({ query: { parent: currentFolder } });
		if (res.error) return;
		const data = res.data as unknown as FolderListResponse;
		subfolders = data.folders;
		folderPhotoCount = data.photo_count;
	}

	async function loadMore() {
		if (loading || !hasMore) return;
		loading = true;

		const query: Record<string, unknown> = { limit: 50 };

		if (isFolderMode) {
			query.sort = 'file_name';
			query.order = 'asc';
			if (currentFolder) query.folder = currentFolder;
		} else {
			query.sort = 'captured_at';
			query.order = 'desc';
		}

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
		if (isFolderMode) loadFolders();
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

	function setGroupBy(mode: 'date' | 'folder') {
		if (groupBy === mode) return;
		groupBy = mode;
		if (mode === 'date') {
			currentFolder = '';
			subfolders = [];
		}
		resetAndLoad();
	}

	function navigateToFolder(path: string) {
		currentFolder = path;
		resetAndLoad();
	}

	function handleFolderClick(name: string) {
		currentFolder = currentFolder ? currentFolder + '/' + name : name;
		resetAndLoad();
	}

	function handlePhotoClick(photo: PhotoListItem) {
		setPhotoNav(photos.map((p) => p.id));
		goto(`/photos/${photo.id}`);
	}

	onMount(() => {
		resetAndLoad();
	});
</script>

<div class="flex h-[calc(100vh-49px)]">
	<!-- Mobile sidebar toggle -->
	<Sheet.Root bind:open={sidebarOpen}>
		<Sheet.Trigger class="fixed bottom-4 left-4 z-30 rounded-full bg-white p-3 shadow-lg border border-gray-200 lg:hidden">
			<svg class="h-5 w-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
			</svg>
			{#if selectedTagIds.length > 0}
				<span class="absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full bg-blue-600 text-[10px] text-white">
					{selectedTagIds.length}
				</span>
			{/if}
		</Sheet.Trigger>
		<Sheet.Content side="left">
			<Sheet.Header>
				<Sheet.Title>Filter by tags</Sheet.Title>
			</Sheet.Header>
			<div class="overflow-y-auto flex-1 p-4">
				<TagSidebar
					{selectedTagIds}
					{tagMode}
					onToggleTag={toggleTag}
					onToggleMode={toggleMode}
					onClear={clearTags}
				/>
			</div>
		</Sheet.Content>
	</Sheet.Root>

	<!-- Desktop sidebar -->
	<aside class="hidden lg:block w-64 overflow-y-auto border-r border-gray-200 bg-white p-4">
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
		<div class="mb-4 flex items-center gap-1 text-sm">
			<span class="text-xs text-gray-500 mr-1">View:</span>
			<button
				type="button"
				onclick={() => setGroupBy('date')}
				class="rounded px-2 py-1 {groupBy === 'date' ? 'bg-gray-200 font-medium text-gray-900' : 'text-gray-500 hover:text-gray-700'}"
			>Timeline</button>
			<button
				type="button"
				onclick={() => setGroupBy('folder')}
				class="rounded px-2 py-1 {groupBy === 'folder' ? 'bg-gray-200 font-medium text-gray-900' : 'text-gray-500 hover:text-gray-700'}"
			>Folders</button>
		</div>

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

		{#if isFolderMode}
			<div class="mb-4">
				<FolderBreadcrumb path={currentFolder} onNavigate={navigateToFolder} />
			</div>
			<div class="mb-4">
				<FolderGrid folders={subfolders} onFolderClick={handleFolderClick} />
			</div>
		{/if}

		<PhotoGrid {photos} {hasMore} {loading} {groupBy} grouped={!isFolderMode} onLoadMore={loadMore} onPhotoClick={handlePhotoClick} />
	</div>
</div>
