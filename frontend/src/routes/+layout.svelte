<script lang="ts">
	import '../app.css';
	import { onMount, setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { AuthState } from '$lib/auth.svelte';
	import { initClient } from '$lib/api/setup';

	let { children } = $props();

	const auth = new AuthState();
	setContext('auth', auth);
	initClient(() => auth.token);

	let ready = $state(false);

	onMount(async () => {
		await auth.restore();
		ready = true;
	});

	const publicPaths = ['/login', '/register'];

	$effect(() => {
		if (!ready) return;
		const path = page.url.pathname;
		if (!auth.isAuthenticated && !publicPaths.includes(path)) {
			goto('/login');
		}
	});

	function handleLogout() {
		auth.logout();
		goto('/login');
	}
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between">
		<a href="/" class="text-xl font-semibold text-gray-900 hover:text-gray-700">HomePhotos</a>

		{#if auth.isAuthenticated}
			<nav class="flex items-center gap-4">
				{#if auth.isAdmin}
					<a href="/admin" class="text-sm text-gray-600 hover:text-gray-900">Admin</a>
				{/if}
				<span class="text-sm text-gray-500">{auth.user?.username}</span>
				<button
					onclick={handleLogout}
					class="text-sm text-gray-600 hover:text-gray-900"
				>
					Sign out
				</button>
			</nav>
		{:else if ready}
			<a href="/login" class="text-sm text-blue-600 hover:underline">Sign in</a>
		{/if}
	</header>

	<main>
		{#if ready}
			{@render children()}
		{:else}
			<div class="flex items-center justify-center p-12">
				<div class="text-gray-400 text-sm">Loading...</div>
			</div>
		{/if}
	</main>
</div>
