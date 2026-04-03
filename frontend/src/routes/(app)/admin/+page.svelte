<script lang="ts">
	import ScannerPanel from '$lib/components/ScannerPanel.svelte';
	import UserManagement from '$lib/components/UserManagement.svelte';
	import TagManagement from '$lib/components/TagManagement.svelte';
	import * as Tabs from '$lib/components/ui/tabs/index.js';

	let activeTab = $state('Scanner');
</script>

<div class="p-6 max-w-4xl mx-auto">
	<h2 class="text-2xl font-semibold text-foreground mb-6">Admin</h2>

	<Tabs.Root bind:value={activeTab}>
		<Tabs.List>
			<Tabs.Trigger value="Scanner">Scanner</Tabs.Trigger>
			<Tabs.Trigger value="Users">Users</Tabs.Trigger>
			<Tabs.Trigger value="Tags">Tags</Tabs.Trigger>
		</Tabs.List>
		<Tabs.Content value="Scanner"><ScannerPanel /></Tabs.Content>
		<Tabs.Content value="Users"><UserManagement /></Tabs.Content>
		<Tabs.Content value="Tags"><TagManagement /></Tabs.Content>
	</Tabs.Root>
</div>
