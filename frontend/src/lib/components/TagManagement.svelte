<script lang="ts">
	import { onMount } from 'svelte';
	import type {
		TagResponse,
		TagListResponse,
		TagGroupResponse,
		TagGroupListResponse
	} from '$lib/api/gen/types.gen';
	import {
		getTags,
		postTag,
		deleteTag,
		getTagGroups,
		postTagGroup,
		deleteTagGroup
	} from '$lib/api/gen/sdk.gen';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';

	let tags: TagResponse[] = $state([]);
	let groups: TagGroupResponse[] = $state([]);

	let newTagName = $state('');
	let newTagColor = $state('#6366f1');
	let newTagGroupId: number | null = $state(null);

	let newGroupName = $state('');
	let newGroupOrder = $state(0);

	let deletingTagId: number | null = $state(null);
	let deletingGroupId: number | null = $state(null);
	let loading = $state(true);
	let error = $state('');

	async function fetchAll() {
		const [tagsRes, groupsRes] = await Promise.all([getTags(), getTagGroups()]);
		if (tagsRes.error || groupsRes.error) {
			error = 'Failed to load tags';
		} else {
			tags = (tagsRes.data as unknown as TagListResponse).data;
			groups = (groupsRes.data as unknown as TagGroupListResponse).data;
			error = '';
		}
		loading = false;
	}

	async function createTag() {
		if (!newTagName.trim()) return;
		error = '';
		const body: { name: string; color?: string; group_id?: number } = { name: newTagName.trim() };
		if (newTagColor) body.color = newTagColor;
		if (newTagGroupId) body.group_id = newTagGroupId;
		const res = await postTag({ body });
		if (res.error) { error = 'Failed to create tag'; return; }
		newTagName = '';
		await fetchAll();
	}

	async function confirmDeleteTag() {
		if (deletingTagId === null) return;
		error = '';
		const res = await deleteTag({ path: { id: deletingTagId } });
		if (res.error) error = 'Failed to delete tag';
		deletingTagId = null;
		await fetchAll();
	}

	async function createGroup() {
		if (!newGroupName.trim()) return;
		error = '';
		const res = await postTagGroup({ body: { name: newGroupName.trim(), sort_order: newGroupOrder } });
		if (res.error) { error = 'Failed to create group'; return; }
		newGroupName = '';
		newGroupOrder = 0;
		await fetchAll();
	}

	async function confirmDeleteGroup() {
		if (deletingGroupId === null) return;
		error = '';
		const res = await deleteTagGroup({ path: { id: deletingGroupId } });
		if (res.error) error = 'Failed to delete group';
		deletingGroupId = null;
		await fetchAll();
	}

	const tagsByGroup = $derived.by(() => {
		const grouped = new Map<string, TagResponse[]>();
		for (const tag of tags) {
			const groupName = tag.group_name ?? 'Ungrouped';
			const existing = grouped.get(groupName);
			if (existing) {
				existing.push(tag);
			} else {
				grouped.set(groupName, [tag]);
			}
		}
		return Array.from(grouped.entries());
	});

	onMount(() => fetchAll());
</script>

<div class="space-y-8">
	{#if error}
		<div class="rounded bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}

	{#if loading}
		<div class="text-sm text-gray-400">Loading...</div>
	{:else}
	<!-- Tag Groups -->
	<div class="space-y-4">
		<h3 class="text-lg font-medium text-gray-900">Tag Groups</h3>

		<div class="flex gap-2">
			<Input type="text" bind:value={newGroupName} placeholder="Group name" />
			<Input type="number" bind:value={newGroupOrder} placeholder="Order" class="w-20" />
			<Button size="sm" onclick={createGroup}>Add</Button>
		</div>

		<div class="space-y-1">
			{#each groups as group (group.id)}
				<div class="flex items-center justify-between rounded border border-gray-200 px-3 py-2">
					<div>
						<span class="text-sm text-gray-900">{group.name}</span>
						<span class="ml-2 text-xs text-gray-400">order: {group.sort_order}</span>
					</div>
					<AlertDialog.Root bind:open={() => deletingGroupId === group.id, (v) => { if (!v) deletingGroupId = null; }}>
						<AlertDialog.Trigger>
							{#snippet child({ props })}
								<Button variant="ghost" size="xs" {...props} class="text-red-500 hover:text-red-700" onclick={() => (deletingGroupId = group.id)}>
									Delete
								</Button>
							{/snippet}
						</AlertDialog.Trigger>
						<AlertDialog.Content>
							<AlertDialog.Header>
								<AlertDialog.Title>Delete group "{group.name}"?</AlertDialog.Title>
								<AlertDialog.Description>This will remove the group. Tags in this group will become ungrouped.</AlertDialog.Description>
							</AlertDialog.Header>
							<AlertDialog.Footer>
								<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
								<AlertDialog.Action onclick={confirmDeleteGroup}>Delete</AlertDialog.Action>
							</AlertDialog.Footer>
						</AlertDialog.Content>
					</AlertDialog.Root>
				</div>
			{/each}
		</div>
	</div>

	<!-- Tags -->
	<div class="space-y-4">
		<h3 class="text-lg font-medium text-gray-900">Tags</h3>

		<div class="flex flex-wrap gap-2">
			<Input type="text" bind:value={newTagName} placeholder="Tag name" />
			<input
				type="color"
				bind:value={newTagColor}
				class="h-8 w-8 cursor-pointer rounded border border-input"
			/>
			<select
				bind:value={newTagGroupId}
				class="rounded-lg border border-input bg-transparent px-2.5 py-1 text-sm"
			>
				<option value={null}>No group</option>
				{#each groups as group (group.id)}
					<option value={group.id}>{group.name}</option>
				{/each}
			</select>
			<Button size="sm" onclick={createTag}>Add</Button>
		</div>

		{#each tagsByGroup as [groupName, groupTags]}
			<div>
				<p class="mb-1 text-xs font-medium uppercase tracking-wider text-gray-400">{groupName}</p>
				<div class="space-y-1">
					{#each groupTags as tag (tag.id)}
						<div class="flex items-center justify-between rounded border border-gray-200 px-3 py-2">
							<div class="flex items-center gap-2">
								{#if tag.color}
									<span class="h-3 w-3 rounded-full" style="background-color: {tag.color}"></span>
								{/if}
								<span class="text-sm text-gray-900">{tag.name}</span>
							</div>
							<AlertDialog.Root bind:open={() => deletingTagId === tag.id, (v) => { if (!v) deletingTagId = null; }}>
								<AlertDialog.Trigger>
									{#snippet child({ props })}
										<Button variant="ghost" size="xs" {...props} class="text-red-500 hover:text-red-700" onclick={() => (deletingTagId = tag.id)}>
											Delete
										</Button>
									{/snippet}
								</AlertDialog.Trigger>
								<AlertDialog.Content>
									<AlertDialog.Header>
										<AlertDialog.Title>Delete tag "{tag.name}"?</AlertDialog.Title>
										<AlertDialog.Description>This will remove the tag from all photos.</AlertDialog.Description>
									</AlertDialog.Header>
									<AlertDialog.Footer>
										<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
										<AlertDialog.Action onclick={confirmDeleteTag}>Delete</AlertDialog.Action>
									</AlertDialog.Footer>
								</AlertDialog.Content>
							</AlertDialog.Root>
						</div>
					{/each}
				</div>
			</div>
		{/each}
	</div>
	{/if}
</div>
