<script lang="ts">
	import { goto } from '$app/navigation';
	import { getContext } from 'svelte';
	import { postAuthRegister } from '$lib/api/gen/sdk.gen';
	import type { AuthResponse } from '$lib/api/gen/types.gen';
	import type { AuthState } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Card from '$lib/components/ui/card/index.js';

	const auth = getContext<AuthState>('auth');

	let username = $state('');
	let password = $state('');
	let email = $state('');
	let error = $state('');
	let submitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		submitting = true;

		const body: { username: string; password: string; email?: string } = { username, password };
		if (email) body.email = email;

		const res = await postAuthRegister({ body });

		if (res.error) {
			const errData = res.error as unknown as { error?: { message?: string } };
			error = errData?.error?.message ?? 'Registration failed';
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
			<Card.Title class="text-2xl">Create account</Card.Title>
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
					<Input id="password" type="password" bind:value={password} required minlength={8} />
					<p class="text-xs text-muted-foreground">Minimum 8 characters</p>
				</div>

				<div class="space-y-1.5">
					<Label for="email">Email (optional)</Label>
					<Input id="email" type="email" bind:value={email} />
				</div>

				<Button type="submit" disabled={submitting} class="w-full">
					{submitting ? 'Creating account...' : 'Create account'}
				</Button>
			</form>

			<p class="mt-4 text-center text-sm text-muted-foreground">
				Already have an account? <a href="/login" class="text-primary hover:underline">Sign in</a>
			</p>
		</Card.Content>
	</Card.Root>
</div>
