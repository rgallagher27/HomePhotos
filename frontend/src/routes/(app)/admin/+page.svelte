<script lang="ts">
	import ScannerPanel from '$lib/components/ScannerPanel.svelte';
	import UserManagement from '$lib/components/UserManagement.svelte';
	import TagManagement from '$lib/components/TagManagement.svelte';

	const tabs = ['Scanner', 'Users', 'Tags'] as const;
	type Tab = (typeof tabs)[number];
	let activeTab: Tab = $state('Scanner');
</script>

<div class="p-6 max-w-4xl mx-auto">
	<h2 class="text-2xl font-semibold text-gray-900 mb-6">Admin</h2>

	<div class="border-b border-gray-200 mb-6">
		<nav class="flex gap-6">
			{#each tabs as tab}
				<button
					type="button"
					onclick={() => (activeTab = tab)}
					class="pb-2 text-sm font-medium border-b-2 transition-colors
					{activeTab === tab
						? 'border-blue-600 text-blue-600'
						: 'border-transparent text-gray-500 hover:text-gray-700'}"
				>
					{tab}
				</button>
			{/each}
		</nav>
	</div>

	{#if activeTab === 'Scanner'}
		<ScannerPanel />
	{:else if activeTab === 'Users'}
		<UserManagement />
	{:else if activeTab === 'Tags'}
		<TagManagement />
	{/if}
</div>
