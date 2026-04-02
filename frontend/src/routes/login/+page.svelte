<script lang="ts">
	import { goto } from '$app/navigation';
	import { getContext } from 'svelte';
	import { postAuthLogin } from '$lib/api/gen/sdk.gen';
	import type { AuthResponse } from '$lib/api/gen/types.gen';
	import type { AuthState } from '$lib/auth.svelte';

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
	<div class="w-full max-w-sm">
		<h2 class="mb-6 text-2xl font-semibold text-gray-900">Sign in</h2>

		{#if error}
			<div class="mb-4 rounded bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
		{/if}

		<form onsubmit={handleSubmit} class="space-y-4">
			<div>
				<label for="username" class="block text-sm font-medium text-gray-700">Username</label>
				<input
					id="username"
					type="text"
					bind:value={username}
					required
					class="mt-1 block w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
				/>
			</div>

			<div>
				<label for="password" class="block text-sm font-medium text-gray-700">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					class="mt-1 block w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
				/>
			</div>

			<button
				type="submit"
				disabled={submitting}
				class="w-full rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
			>
				{submitting ? 'Signing in...' : 'Sign in'}
			</button>
		</form>

		<p class="mt-4 text-center text-sm text-gray-600">
			Don't have an account? <a href="/register" class="text-blue-600 hover:underline">Register</a>
		</p>
	</div>
</div>
