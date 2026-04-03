<script lang="ts">
	import { onMount } from 'svelte';
	import type { UserResponse, UserListResponse } from '$lib/api/gen/types.gen';
	import { getUsers, putUserRole } from '$lib/api/gen/sdk.gen';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';

	let users: UserResponse[] = $state([]);
	let confirmingId: number | null = $state(null);
	let pendingRole: 'admin' | 'viewer' | null = $state(null);
	let loading = $state(true);
	let error = $state('');

	async function fetchUsers() {
		const res = await getUsers();
		if (res.error) {
			error = 'Failed to load users';
		} else {
			users = (res.data as unknown as UserListResponse).data;
			error = '';
		}
		loading = false;
	}

	function requestRoleChange(userId: number, role: 'admin' | 'viewer') {
		confirmingId = userId;
		pendingRole = role;
	}

	async function confirmRoleChange() {
		if (confirmingId === null || !pendingRole) return;
		const res = await putUserRole({ path: { id: confirmingId }, body: { role: pendingRole } });
		if (res.error) error = 'Failed to update role';
		confirmingId = null;
		pendingRole = null;
		await fetchUsers();
	}

	function cancelRoleChange() {
		confirmingId = null;
		pendingRole = null;
	}

	function formatDate(dateStr: string | null | undefined): string {
		if (!dateStr) return '—';
		return new Date(dateStr).toLocaleDateString(undefined, {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	onMount(() => fetchUsers());
</script>

<div class="space-y-4">
	<h3 class="text-lg font-medium text-gray-900">Users</h3>

	{#if error}
		<div class="rounded bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}

	{#if loading}
		<div class="space-y-2">
			{#each { length: 4 } as _}
				<Skeleton class="h-10 w-full" />
			{/each}
		</div>
	{:else}
	<AlertDialog.Root bind:open={() => confirmingId !== null, (v) => { if (!v) cancelRoleChange(); }}>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>Change role to {pendingRole}?</AlertDialog.Title>
				<AlertDialog.Description>This will change the user's permissions.</AlertDialog.Description>
			</AlertDialog.Header>
			<AlertDialog.Footer>
				<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
				<AlertDialog.Action onclick={confirmRoleChange}>Confirm</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>

	<Table.Root>
		<Table.Header>
			<Table.Row>
				<Table.Head>Username</Table.Head>
				<Table.Head>Role</Table.Head>
				<Table.Head>Created</Table.Head>
				<Table.Head>Last Login</Table.Head>
			</Table.Row>
		</Table.Header>
		<Table.Body>
			{#each users as user (user.id)}
				<Table.Row>
					<Table.Cell>{user.username}</Table.Cell>
					<Table.Cell>
						<select
							value={user.role}
							onchange={(e) => requestRoleChange(user.id, (e.target as HTMLSelectElement).value as 'admin' | 'viewer')}
							class="rounded border border-input px-2 py-1 text-xs"
						>
							<option value="admin">admin</option>
							<option value="viewer">viewer</option>
						</select>
					</Table.Cell>
					<Table.Cell class="text-muted-foreground">{formatDate(user.created_at)}</Table.Cell>
					<Table.Cell class="text-muted-foreground">{formatDate(user.last_login)}</Table.Cell>
				</Table.Row>
			{/each}
		</Table.Body>
	</Table.Root>
	{/if}
</div>
