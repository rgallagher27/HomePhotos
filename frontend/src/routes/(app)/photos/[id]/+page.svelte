<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import type { PhotoDetailResponse } from '$lib/api/gen/types.gen';
	import { getPhoto } from '$lib/api/gen/sdk.gen';
	import { previewUrl, fullUrl } from '$lib/image';
	import ExifPanel from '$lib/components/ExifPanel.svelte';
	import PhotoTags from '$lib/components/PhotoTags.svelte';

	let photo: PhotoDetailResponse | null = $state(null);
	let error = $state('');
	let showFull = $state(false);

	const photoId = $derived(Number(page.params.id));

	async function loadPhoto() {
		const res = await getPhoto({ path: { id: photoId } });
		if (res.error) {
			error = 'Photo not found';
			return;
		}
		photo = res.data as unknown as PhotoDetailResponse;
	}

	onMount(() => {
		loadPhoto();
	});
</script>

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
				<div class="relative rounded-lg bg-gray-100 overflow-hidden">
					<img
						src={showFull ? fullUrl(photo.id) : previewUrl(photo.id)}
						alt={photo.file_name}
						class="mx-auto max-h-[80vh] object-contain"
					/>
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
