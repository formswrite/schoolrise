<script lang="ts">
	import type { PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';

	let { data }: { data: PageData } = $props();
</script>

<div class="space-y-6">
	<header class="flex items-center justify-between">
		<h1 class="text-2xl font-semibold">Users</h1>
		<Button href="/admin/users/new">+ New user</Button>
	</header>

	{#if data.users.length === 0}
		<Card.Root>
			<Card.Content class="py-12 text-center text-sm text-muted-foreground">
				No users yet. Click <strong>+ New user</strong> to add one.
			</Card.Content>
		</Card.Root>
	{:else}
		<Card.Root>
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Email</Table.Head>
						<Table.Head>Full name</Table.Head>
						<Table.Head>Role</Table.Head>
						<Table.Head>Last login</Table.Head>
						<Table.Head>Status</Table.Head>
						<Table.Head class="w-px"></Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.users as user}
						<Table.Row>
							<Table.Cell class="font-medium">{user.email}</Table.Cell>
							<Table.Cell>{user.fullName}</Table.Cell>
							<Table.Cell>{user.role}</Table.Cell>
							<Table.Cell class="text-muted-foreground">
								{user.lastLoginAt ? new Date(user.lastLoginAt).toLocaleString() : '—'}
							</Table.Cell>
							<Table.Cell>
								{#if user.lockedAt}
									<Badge variant="destructive">Locked</Badge>
								{:else if user.mustChangePassword}
									<Badge variant="secondary">Pwd change</Badge>
								{:else}
									<Badge>Active</Badge>
								{/if}
							</Table.Cell>
							<Table.Cell>
								<Button href="/admin/users/{user.id}" variant="ghost" size="sm" class="h-7 px-2 text-xs">Manage</Button>
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</Card.Root>
	{/if}
</div>
