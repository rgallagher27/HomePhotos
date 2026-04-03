<script lang="ts">
	import { goto } from '$app/navigation';
	import { getContext } from 'svelte';
	import { postAuthLogin } from '$lib/api/gen/sdk.gen';
	import type { AuthResponse } from '$lib/api/gen/types.gen';
	import type { AuthState } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Card from '$lib/components/ui/card/index.js';

	const auth = getContext<AuthState>('auth');

	let username = $state('');
	let password = $state('');
	let error = $state('');
	let submitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		submitting = true;

		const res = await postAuthLogin({ body: { username, password } });

		if (res.error) {
			error = 'Invalid username or password';
			submitting = false;
			return;
		}

		auth.login(res.data as unknown as AuthResponse);
		goto('/');
	}
</script>

<div class="flex min-h-[80vh] items-center justify-center">
	<Card.Root class="w-full max-w-sm">
		<Card.Header>
			<Card.Title class="text-2xl">Sign in</Card.Title>
		</Card.Header>
		<Card.Content>
			{#if error}
				<div class="mb-4 rounded bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
			{/if}

			<form onsubmit={handleSubmit} class="space-y-4">
				<div class="space-y-1.5">
					<Label for="username">Username</Label>
					<Input id="username" type="text" bind:value={username} required />
				</div>

				<div class="space-y-1.5">
					<Label for="password">Password</Label>
					<Input id="password" type="password" bind:value={password} required />
				</div>

				<Button type="submit" disabled={submitting} class="w-full">
					{submitting ? 'Signing in...' : 'Sign in'}
				</Button>
			</form>

			<p class="mt-4 text-center text-sm text-muted-foreground">
				Don't have an account? <a href="/register" class="text-primary hover:underline">Register</a>
			</p>
		</Card.Content>
	</Card.Root>
</div>
