<script lang="ts">
	import { onMount } from 'svelte';
	import type {
		TagResponse,
		TagListResponse,
		TagGroupResponse,
		TagGroupListResponse
	} from '$lib/api/gen/types.gen';
	import { getTags, getTagGroups } from '$lib/api/gen/sdk.gen';
	import TagChip from './TagChip.svelte';

	let {
		selectedTagIds,
		tagMode,
		onToggleTag,
		onToggleMode,
		onClear
	}: {
		selectedTagIds: number[];
		tagMode: 'and' | 'or';
		onToggleTag: (id: number) => void;
		onToggleMode: () => void;
		onClear: () => void;
	} = $props();

	let allTags: TagResponse[] = $state([]);
	let groups: TagGroupResponse[] = $state([]);

	onMount(async () => {
		const [tagsRes, groupsRes] = await Promise.all([getTags(), getTagGroups()]);
		if (tagsRes.data) allTags = (tagsRes.data as unknown as TagListResponse).data;
		if (groupsRes.data) groups = (groupsRes.data as unknown as TagGroupListResponse).data;
	});

	const selectedSet = $derived(new Set(selectedTagIds));

	type TagGroup = { name: string; tags: TagResponse[] };

	const groupedTags: TagGroup[] = $derived.by(() => {
		const groupMap = new Map<number, TagGroup>();
		for (const g of groups) {
			groupMap.set(g.id, { name: g.name, tags: [] });
		}

		const ungrouped: TagResponse[] = [];
		for (const tag of allTags) {
			if (tag.group_id && groupMap.has(tag.group_id)) {
				groupMap.get(tag.group_id)!.tags.push(tag);
			} else {
				ungrouped.push(tag);
			}
		}

		const result: TagGroup[] = [];
		for (const g of groups) {
			const group = groupMap.get(g.id)!;
			if (group.tags.length > 0) result.push(group);
		}
		if (ungrouped.length > 0) result.push({ name: 'Other', tags: ungrouped });
		return result;
	});
</script>

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<h3 class="text-sm font-medium text-gray-900">Filter by Tags</h3>
		{#if selectedTagIds.length > 0}
			<button type="button" onclick={onClear} class="text-xs text-gray-400 hover:text-gray-600">
				Clear
			</button>
		{/if}
	</div>

	{#if allTags.length > 0}
		<div class="flex items-center gap-2">
			<span class="text-xs text-gray-500">Mode:</span>
			<button
				type="button"
				onclick={onToggleMode}
				class="rounded px-2 py-0.5 text-xs font-medium {tagMode === 'or'
					? 'bg-blue-100 text-blue-700'
					: 'bg-gray-100 text-gray-600'}"
			>
				Any
			</button>
			<button
				type="button"
				onclick={onToggleMode}
				class="rounded px-2 py-0.5 text-xs font-medium {tagMode === 'and'
					? 'bg-blue-100 text-blue-700'
					: 'bg-gray-100 text-gray-600'}"
			>
				All
			</button>
		</div>
	{/if}

	{#each groupedTags as group}
		<div>
			<p class="mb-1.5 text-xs font-medium text-gray-400 uppercase tracking-wider">
				{group.name}
			</p>
			<div class="flex flex-wrap gap-1.5">
				{#each group.tags as tag (tag.id)}
					<TagChip
						name={tag.name}
						color={tag.color}
						selected={selectedSet.has(tag.id)}
						onclick={() => onToggleTag(tag.id)}
					/>
				{/each}
			</div>
		</div>
	{/each}

	{#if allTags.length === 0}
		<p class="text-xs text-gray-400">No tags yet</p>
	{/if}
</div>
