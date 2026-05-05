<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Badge } from '$lib/components/ui/badge';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	function scopeLabel(scopeNodeId: number | null): string {
		return scopeNodeId === null ? 'Global' : `Node #${scopeNodeId}`;
	}
</script>

<div class="space-y-6">
	<header>
		<a href="/admin/users" class="text-xs text-muted-foreground hover:underline">← All users</a>
		<h1 class="mt-1 text-2xl font-semibold">{data.user.fullName}</h1>
		<p class="text-sm text-muted-foreground">{data.user.email}</p>
	</header>

	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">Account</Card.Title>
		</Card.Header>
		<Card.Content>
			<dl class="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
				<dt class="text-muted-foreground">User ID</dt>
				<dd>{data.user.id}</dd>
				<dt class="text-muted-foreground">Role label</dt>
				<dd>{data.user.role}</dd>
				<dt class="text-muted-foreground">Created</dt>
				<dd>{new Date(data.user.createdAt).toLocaleString()}</dd>
				<dt class="text-muted-foreground">Last login</dt>
				<dd>{data.user.lastLoginAt ? new Date(data.user.lastLoginAt).toLocaleString() : '—'}</dd>
				<dt class="text-muted-foreground">Status</dt>
				<dd>
					{#if data.user.lockedAt}
						<Badge variant="destructive">Locked</Badge>
					{:else if data.user.mustChangePassword}
						<Badge variant="secondary">Must change password</Badge>
					{:else}
						<Badge>Active</Badge>
					{/if}
				</dd>
			</dl>
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">Role assignments</Card.Title>
			<Card.Description>
				Each assignment grants the user a role at a specific scope. <strong>Global</strong> means no scope restriction (full access).
			</Card.Description>
		</Card.Header>
		<Card.Content class="space-y-6">
			{#if data.assignments.length === 0}
				<Alert>
					<AlertDescription>
						This user has no role assignments and cannot access any data. Add one below.
					</AlertDescription>
				</Alert>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Role</Table.Head>
							<Table.Head>Scope</Table.Head>
							<Table.Head class="w-px"></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each data.assignments as a}
							<Table.Row>
								<Table.Cell class="font-medium">{a.role}</Table.Cell>
								<Table.Cell>{scopeLabel(a.scopeNodeId)}</Table.Cell>
								<Table.Cell>
									<form method="POST" action="?/revoke">
										<input type="hidden" name="assignment_id" value={a.id} />
										<Button
											type="submit"
											variant="ghost"
											size="sm"
											class="h-7 px-2 text-xs text-destructive hover:text-destructive"
											onclick={(e) => { if (!confirm('Revoke this assignment?')) e.preventDefault(); }}
										>
											Revoke
										</Button>
									</form>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}

			<form method="POST" action="?/assign" class="rounded-md border bg-muted/30 p-4">
				<h3 class="text-sm font-medium">Assign role</h3>
				<div class="mt-3 grid grid-cols-3 gap-3">
					<div class="space-y-1.5">
						<Label for="role" class="text-xs">Role</Label>
						<select
							id="role"
							name="role"
							required
							class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs focus:outline-none focus:ring-1 focus:ring-ring"
						>
							<option value="admin">admin</option>
							<option value="inspector" selected>inspector</option>
							<option value="teacher">teacher</option>
						</select>
					</div>
					<div class="col-span-2 space-y-1.5">
						<Label for="scope_node_id" class="text-xs">Scope (node ID)</Label>
						<Input id="scope_node_id" name="scope_node_id" type="number" min="1" placeholder="Leave blank for global" />
						<p class="text-xs text-muted-foreground">
							Find a node ID from <a href="/admin/institutions" class="underline">/admin/institutions</a> (URL has <code>?parent=NN</code>).
						</p>
					</div>
				</div>

				{#if form?.error}
					<Alert variant="destructive" class="mt-3">
						<AlertDescription>{form.error}</AlertDescription>
					</Alert>
				{/if}

				<div class="mt-3 flex justify-end">
					<Button type="submit" size="sm">Assign</Button>
				</div>
			</form>
		</Card.Content>
	</Card.Root>
</div>
