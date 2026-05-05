<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Badge } from '$lib/components/ui/badge';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import FormToaster from '$lib/components/form-toaster.svelte';
	import { enhance } from '$app/forms';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	let showForm = $state(false);
</script>

<div class="space-y-6">
	<FormToaster {form} />
	<header class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold">Academic periods</h1>
			<p class="mt-1 text-sm text-muted-foreground">
				School years or terms. Exactly one is "current" at any time.
			</p>
		</div>
		<Button size="sm" onclick={() => (showForm = !showForm)}>
			{showForm ? 'Cancel' : '+ New period'}
		</Button>
	</header>

	{#if showForm}
		<Card.Root>
			<Card.Content class="pt-6">
				<form method="POST" use:enhance action="?/create" class="space-y-4">
					<div class="grid grid-cols-2 gap-4">
						<div class="space-y-2">
							<Label for="code">Code *</Label>
							<Input id="code" name="code" required placeholder="2025-2026" />
						</div>
						<div class="space-y-2">
							<Label for="label">Label *</Label>
							<Input id="label" name="label" required placeholder="School Year 2025-2026" />
						</div>
					</div>
					<div class="grid grid-cols-2 gap-4">
						<div class="space-y-2">
							<Label for="starts_on">Starts on *</Label>
							<Input id="starts_on" name="starts_on" type="date" required />
						</div>
						<div class="space-y-2">
							<Label for="ends_on">Ends on *</Label>
							<Input id="ends_on" name="ends_on" type="date" required />
						</div>
					</div>
					<div class="flex items-center gap-2">
						<Checkbox id="is_current" name="is_current" />
						<Label for="is_current" class="cursor-pointer">Mark as current period</Label>
					</div>
					{#if form?.error}
						<Alert variant="destructive">
							<AlertDescription>{form.error}</AlertDescription>
						</Alert>
					{/if}
					<div class="flex justify-end">
						<Button type="submit" size="sm">Create</Button>
					</div>
				</form>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if data.periods.length === 0}
		<Card.Root>
			<Card.Content class="py-12 text-center text-sm text-muted-foreground">
				No academic periods yet.
			</Card.Content>
		</Card.Root>
	{:else}
		<Card.Root>
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Code</Table.Head>
						<Table.Head>Label</Table.Head>
						<Table.Head>Starts</Table.Head>
						<Table.Head>Ends</Table.Head>
						<Table.Head>Status</Table.Head>
						<Table.Head class="w-px"></Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.periods as p}
						<Table.Row>
							<Table.Cell class="font-medium">{p.code}</Table.Cell>
							<Table.Cell>{p.label}</Table.Cell>
							<Table.Cell class="text-muted-foreground">{new Date(p.starts_on).toLocaleDateString()}</Table.Cell>
							<Table.Cell class="text-muted-foreground">{new Date(p.ends_on).toLocaleDateString()}</Table.Cell>
							<Table.Cell>
								{#if p.is_current}
									<Badge>Current</Badge>
								{:else}
									<Badge variant="secondary">Inactive</Badge>
								{/if}
							</Table.Cell>
							<Table.Cell>
								<div class="flex items-center gap-1">
									{#if !p.is_current}
										<form method="POST" use:enhance action="?/setCurrent">
											<input type="hidden" name="id" value={p.id} />
											<Button type="submit" variant="ghost" size="sm" class="h-7 px-2 text-xs">Make current</Button>
										</form>
									{/if}
									<form method="POST" use:enhance action="?/delete">
										<input type="hidden" name="id" value={p.id} />
										<Button
											type="submit"
											variant="ghost"
											size="sm"
											class="h-7 px-2 text-xs text-destructive hover:text-destructive"
											onclick={(e) => { if (!confirm('Delete this period?')) e.preventDefault(); }}
										>
											Delete
										</Button>
									</form>
								</div>
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</Card.Root>
	{/if}
</div>
