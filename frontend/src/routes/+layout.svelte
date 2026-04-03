<script lang="ts">
	import '../app.css';
	import { onMount, setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { AuthState } from '$lib/auth.svelte';
	import { initClient } from '$lib/api/setup';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Spinner } from '$lib/components/ui/spinner/index.js';

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

<div class="min-h-screen bg-background">
	<header class="bg-card border-b border-border px-6 py-3 flex items-center justify-between">
		<a href="/" class="text-xl font-semibold text-foreground hover:text-foreground/80">HomePhotos</a>

		{#if auth.isAuthenticated}
			<nav class="flex items-center gap-2">
				{#if auth.isAdmin}
					<Button variant="ghost" size="sm" href="/admin">Admin</Button>
				{/if}
				<span class="text-sm text-muted-foreground">{auth.user?.username}</span>
				<Button variant="ghost" size="sm" onclick={handleLogout}>Sign out</Button>
			</nav>
		{:else if ready}
			<Button variant="link" href="/login">Sign in</Button>
		{/if}
	</header>

	<main>
		{#if ready}
			{@render children()}
		{:else}
			<div class="flex items-center justify-center p-12">
				<Spinner />
			</div>
		{/if}
	</main>
</div>
