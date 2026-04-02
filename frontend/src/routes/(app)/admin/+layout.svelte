<script lang="ts">
	import { goto } from '$app/navigation';
	import { getContext } from 'svelte';
	import type { AuthState } from '$lib/auth.svelte';

	let { children } = $props();
	const auth = getContext<AuthState>('auth');

	$effect(() => {
		if (!auth.isAdmin) goto('/');
	});
</script>

{#if auth.isAdmin}
	{@render children()}
{/if}
