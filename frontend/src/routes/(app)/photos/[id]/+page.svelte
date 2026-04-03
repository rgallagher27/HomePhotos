<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { PhotoDetailResponse } from '$lib/api/gen/types.gen';
	import { getPhoto } from '$lib/api/gen/sdk.gen';
	import { previewUrl, fullUrl } from '$lib/image';
	import ExifPanel from '$lib/components/ExifPanel.svelte';
	import PhotoTags from '$lib/components/PhotoTags.svelte';
	import { getAdjacentId } from '$lib/photoNav.svelte';

	let photo: PhotoDetailResponse | null = $state(null);
	let error = $state('');
	let showFull = $state(false);

	const photoId = $derived(Number(page.params.id));
	const prevId = $derived(getAdjacentId(photoId, 'prev'));
	const nextId = $derived(getAdjacentId(photoId, 'next'));

	async function loadPhoto() {
		const res = await getPhoto({ path: { id: photoId } });
		if (res.error) {
			error = 'Photo not found';
			return;
		}
		photo = res.data as unknown as PhotoDetailResponse;
	}

	function navigate(id: number | null) {
		if (id == null) return;
		showFull = false;
		photo = null;
		goto(`/photos/${id}`);
	}

	function handleKeydown(e: KeyboardEvent) {
		if ((e.target as HTMLElement).tagName === 'INPUT') return;
		if (e.key === 'ArrowLeft') navigate(prevId);
		else if (e.key === 'ArrowRight') navigate(nextId);
	}

	$effect(() => {
		// Load photo on mount and when photoId changes (arrow key navigation)
		// Access photoId to create the dependency
		const _id = photoId;
		loadPhoto();
	});
</script>

<svelte:window onkeydown={handleKeydown} />

{#if error}
	<div class="flex items-center justify-center p-12">
		<div class="text-center">
			<p class="text-gray-500">{error}</p>
			<a href="/" class="mt-2 inline-block text-sm text-blue-600 hover:underline">Back to photos</a>
		</div>
	</div>
{:else if !photo}
	<div class="flex items-center justify-center p-12">
		<div class="text-sm text-gray-400">Loading...</div>
	</div>
{:else}
	<div class="p-4">
		<a href="/" class="mb-4 inline-flex items-center text-sm text-gray-500 hover:text-gray-700">
			&larr; Back
		</a>

		<div class="flex flex-col gap-6 lg:flex-row">
			<div class="flex-1 min-w-0">
				<div class="relative rounded-lg bg-gray-100 overflow-hidden group">
					{#if prevId != null}
						<button
							type="button"
							onclick={() => navigate(prevId)}
							class="absolute left-2 top-1/2 -translate-y-1/2 z-10 rounded-full bg-black/40 p-2 text-white opacity-0 transition-opacity group-hover:opacity-100 hover:bg-black/60"
							aria-label="Previous photo"
						>
							<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
						</button>
					{/if}
					<img
						src={showFull ? fullUrl(photo.id) : previewUrl(photo.id)}
						alt={photo.file_name}
						class="mx-auto max-h-[80vh] object-contain"
					/>
					{#if nextId != null}
						<button
							type="button"
							onclick={() => navigate(nextId)}
							class="absolute right-2 top-1/2 -translate-y-1/2 z-10 rounded-full bg-black/40 p-2 text-white opacity-0 transition-opacity group-hover:opacity-100 hover:bg-black/60"
							aria-label="Next photo"
						>
							<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
						</button>
					{/if}
				</div>
				<div class="mt-2 flex items-center justify-between">
					<h2 class="truncate text-sm font-medium text-gray-700">{photo.file_name}</h2>
					<button
						type="button"
						onclick={() => (showFull = !showFull)}
						class="text-xs text-blue-600 hover:underline whitespace-nowrap ml-2"
					>
						{showFull ? 'Show preview' : 'View full resolution'}
					</button>
				</div>
			</div>

			<div class="w-full lg:w-72 space-y-6 shrink-0">
				<PhotoTags photoId={photo.id} tags={photo.tags ?? []} onUpdate={loadPhoto} />
				<ExifPanel {photo} />
			</div>
		</div>
	</div>
{/if}
