<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import ScopePicker from '$lib/components/scope-picker.svelte';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<div class="space-y-6">
	<header class="space-y-3">
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-semibold">Staff</h1>
				{#if data.node}
					<p class="mt-1 text-sm text-muted-foreground">
						at <a href="/admin/institutions?parent={data.node.id}" class="font-medium text-foreground hover:underline">{data.node.label}</a>
					</p>
				{/if}
			</div>
			{#if data.scopeNodeId}
				<Button href="/admin/staff/new?scope={data.scopeNodeId}">+ New staff</Button>
			{/if}
		</div>
		{#if data.scopeOptions && data.scopeOptions.length > 0 && data.scopeNodeId}
			<ScopePicker options={data.scopeOptions} current={data.scopeNodeId} />
		{/if}
	</header>

	{#if !data.scopeNodeId}
		<Alert>
			<AlertDescription>
				No scopes are configured yet. Set up your hierarchy first via <a href="/admin/institutions" class="font-medium underline">Institutions</a>.
			</AlertDescription>
		</Alert>
	{:else if data.staff.length === 0}
		<Card.Root>
			<Card.Content class="py-12 text-center text-sm text-muted-foreground">
				No staff at this scope yet.
			</Card.Content>
		</Card.Root>
	{:else}
		<Card.Root>
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Name</Table.Head>
						<Table.Head>Position</Table.Head>
						<Table.Head>Code</Table.Head>
						<Table.Head>Hired</Table.Head>
						<Table.Head>Email</Table.Head>
						<Table.Head class="w-px"></Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.staff as person}
						<Table.Row>
							<Table.Cell class="font-medium">{person.person.fullName}</Table.Cell>
							<Table.Cell>{person.position || '—'}</Table.Cell>
							<Table.Cell>{person.staffCode || '—'}</Table.Cell>
							<Table.Cell class="text-muted-foreground">
								{person.hireDate ? new Date(person.hireDate).toLocaleDateString() : '—'}
							</Table.Cell>
							<Table.Cell class="text-muted-foreground">{person.person.email || '—'}</Table.Cell>
							<Table.Cell>
								<form method="POST" action="?/delete">
									<input type="hidden" name="id" value={person.id} />
									<Button
										type="submit"
										variant="ghost"
										size="sm"
										class="h-7 px-2 text-xs text-destructive hover:text-destructive"
										onclick={(e) => { if (!confirm('Delete this staff member?')) e.preventDefault(); }}
									>
										Delete
									</Button>
								</form>
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</Card.Root>

		{#if form?.error}
			<Alert variant="destructive">
				<AlertDescription>{form.error}</AlertDescription>
			</Alert>
		{/if}
	{/if}
</div>
