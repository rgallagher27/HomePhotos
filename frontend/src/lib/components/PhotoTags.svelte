<script lang="ts">
	import { getContext } from 'svelte';
	import type { TagResponse, TagListResponse } from '$lib/api/gen/types.gen';
	import { getTags, postTag, postPhotoTags, deletePhotoTag } from '$lib/api/gen/sdk.gen';
	import type { AuthState } from '$lib/auth.svelte';

	let {
		photoId,
		tags,
		onUpdate
	}: {
		photoId: number;
		tags: TagResponse[];
		onUpdate: () => void;
	} = $props();

	const auth = getContext<AuthState>('auth');

	let allTags: TagResponse[] = $state([]);
	let showDropdown = $state(false);
	let search = $state('');
	let adding = $state(false);
	let error = $state('');

	async function loadTags() {
		const res = await getTags();
		if (res.error) {
			error = 'Failed to load tags';
			return;
		}
		allTags = (res.data as unknown as TagListResponse).data;
	}

	function openDropdown() {
		showDropdown = true;
		search = '';
		loadTags();
	}

	const currentTagIds = $derived(new Set(tags.map((t) => t.id)));

	const filtered = $derived(
		allTags
			.filter((t) => !currentTagIds.has(t.id))
			.filter((t) => !search || t.name.toLowerCase().includes(search.toLowerCase()))
	);

	const showCreate = $derived(
		auth.isAdmin &&
			search.trim() !== '' &&
			!allTags.some((t) => t.name.toLowerCase() === search.trim().toLowerCase())
	);

	async function addTag(tagId: number) {
		adding = true;
		error = '';
		const res = await postPhotoTags({ path: { id: photoId }, body: { tag_ids: [tagId] } });
		if (res.error) error = 'Failed to add tag';
		showDropdown = false;
		adding = false;
		onUpdate();
	}

	async function createAndAdd() {
		adding = true;
		error = '';
		const res = await postTag({ body: { name: search.trim() } });
		if (res.error) {
			error = 'Failed to create tag';
			adding = false;
			return;
		}
		const created = res.data as unknown as TagResponse;
		const assignRes = await postPhotoTags({ path: { id: photoId }, body: { tag_ids: [created.id] } });
		if (assignRes.error) error = 'Failed to assign tag';
		showDropdown = false;
		adding = false;
		onUpdate();
	}

	async function removeTag(tagId: number) {
		error = '';
		const res = await deletePhotoTag({ path: { id: photoId, tagId } });
		if (res.error) error = 'Failed to remove tag';
		onUpdate();
	}

	function tagBg(color: string | null | undefined): string {
		if (!color) return 'bg-gray-100 text-gray-700';
		return `text-white`;
	}
</script>

<div class="space-y-2">
	<h3 class="text-sm font-medium text-gray-900">Tags</h3>

	{#if error}
		<div class="rounded bg-red-50 px-3 py-2 text-xs text-red-700">{error}</div>
	{/if}

	<div class="flex flex-wrap gap-1.5">
		{#each tags as tag (tag.id)}
			<span
				class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium {tagBg(tag.color)}"
				style={tag.color ? `background-color: ${tag.color}` : ''}
			>
				{tag.name}
				<button
					type="button"
					onclick={() => removeTag(tag.id)}
					class="ml-0.5 hover:opacity-70"
					aria-label="Remove tag {tag.name}"
				>
					×
				</button>
			</span>
		{/each}

		{#if !showDropdown}
			<button
				type="button"
				onclick={openDropdown}
				class="inline-flex items-center rounded-full border border-dashed border-gray-300 px-2.5 py-0.5 text-xs text-gray-500 hover:border-gray-400 hover:text-gray-700"
			>
				+ Add tag
			</button>
		{/if}
	</div>

	{#if showDropdown}
		<div class="relative">
			<input
				type="text"
				bind:value={search}
				placeholder="Search tags..."
				disabled={adding}
				class="w-full rounded border border-gray-300 px-3 py-1.5 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
			/>
			<div
				role="listbox"
				tabindex="-1"
				aria-label="Available tags"
				class="absolute z-10 mt-1 max-h-48 w-full overflow-auto rounded border border-gray-200 bg-white shadow-lg"
				onmousedown={(e) => e.preventDefault()}
			>
				{#each filtered as tag (tag.id)}
					<button
						type="button"
						onclick={() => addTag(tag.id)}
						disabled={adding}
						class="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm hover:bg-gray-50 disabled:opacity-50"
					>
						{#if tag.color}
							<span class="h-3 w-3 rounded-full" style="background-color: {tag.color}"></span>
						{/if}
						{tag.name}
					</button>
				{/each}

				{#if showCreate}
					<button
						type="button"
						onclick={createAndAdd}
						disabled={adding}
						class="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-blue-600 hover:bg-blue-50 disabled:opacity-50"
					>
						Create "{search.trim()}"
					</button>
				{/if}

				{#if filtered.length === 0 && !showCreate}
					<div class="px-3 py-2 text-sm text-gray-400">No tags available</div>
				{/if}
			</div>

			<button
				type="button"
				onclick={() => (showDropdown = false)}
				class="mt-1 text-xs text-gray-400 hover:text-gray-600"
			>
				Cancel
			</button>
		</div>
	{/if}
</div>
