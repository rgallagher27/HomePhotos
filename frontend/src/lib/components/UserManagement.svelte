<script lang="ts">
	import { onMount } from 'svelte';
	import type { UserResponse, UserListResponse } from '$lib/api/gen/types.gen';
	import { getUsers, putUserRole } from '$lib/api/gen/sdk.gen';

	let users: UserResponse[] = $state([]);
	let confirmingId: number | null = $state(null);
	let pendingRole: 'admin' | 'viewer' | null = $state(null);

	async function fetchUsers() {
		const res = await getUsers();
		if (res.data) {
			users = (res.data as unknown as UserListResponse).data;
		}
	}

	function requestRoleChange(userId: number, role: 'admin' | 'viewer') {
		confirmingId = userId;
		pendingRole = role;
	}

	async function confirmRoleChange() {
		if (confirmingId === null || !pendingRole) return;
		await putUserRole({ path: { id: confirmingId }, body: { role: pendingRole } });
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

	<div class="overflow-x-auto rounded border border-gray-200">
		<table class="min-w-full divide-y divide-gray-200">
			<thead class="bg-gray-50">
				<tr>
					<th class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500">Username</th>
					<th class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500">Role</th>
					<th class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500">Created</th>
					<th class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500">Last Login</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-gray-200">
				{#each users as user (user.id)}
					<tr>
						<td class="whitespace-nowrap px-4 py-2 text-sm text-gray-900">{user.username}</td>
						<td class="whitespace-nowrap px-4 py-2 text-sm">
							{#if confirmingId === user.id}
								<span class="text-amber-600 text-xs">
									Change to {pendingRole}?
									<button type="button" onclick={confirmRoleChange} class="ml-1 font-medium text-blue-600 hover:underline">Yes</button>
									<button type="button" onclick={cancelRoleChange} class="ml-1 font-medium text-gray-400 hover:underline">No</button>
								</span>
							{:else}
								<select
									value={user.role}
									onchange={(e) => requestRoleChange(user.id, (e.target as HTMLSelectElement).value as 'admin' | 'viewer')}
									class="rounded border border-gray-300 px-2 py-1 text-xs"
								>
									<option value="admin">admin</option>
									<option value="viewer">viewer</option>
								</select>
							{/if}
						</td>
						<td class="whitespace-nowrap px-4 py-2 text-sm text-gray-500">{formatDate(user.created_at)}</td>
						<td class="whitespace-nowrap px-4 py-2 text-sm text-gray-500">{formatDate(user.last_login)}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>
