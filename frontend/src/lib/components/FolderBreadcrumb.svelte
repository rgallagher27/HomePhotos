<script lang="ts">
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';

	let {
		path,
		onNavigate
	}: {
		path: string;
		onNavigate: (path: string) => void;
	} = $props();

	const segments: { label: string; path: string }[] = $derived.by(() => {
		if (!path) return [];
		const parts = path.split('/');
		return parts.map((part, i) => ({
			label: part,
			path: parts.slice(0, i + 1).join('/')
		}));
	});
</script>

<Breadcrumb.Root>
	<Breadcrumb.List>
		<Breadcrumb.Item>
			{#if path}
				<Breadcrumb.Link onclick={() => onNavigate('')} class="cursor-pointer">All folders</Breadcrumb.Link>
			{:else}
				<Breadcrumb.Page>All folders</Breadcrumb.Page>
			{/if}
		</Breadcrumb.Item>
		{#each segments as segment (segment.path)}
			<Breadcrumb.Separator />
			<Breadcrumb.Item>
				{#if segment.path === path}
					<Breadcrumb.Page>{segment.label}</Breadcrumb.Page>
				{:else}
					<Breadcrumb.Link onclick={() => onNavigate(segment.path)} class="cursor-pointer">{segment.label}</Breadcrumb.Link>
				{/if}
			</Breadcrumb.Item>
		{/each}
	</Breadcrumb.List>
</Breadcrumb.Root>
