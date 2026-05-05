<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
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
			<h1 class="text-2xl font-semibold">Niveaux (grade levels)</h1>
			<p class="mt-1 text-sm text-muted-foreground">
				Define the curricular progression — e.g. CE1, CE2, CM1.
			</p>
		</div>
		<Button size="sm" onclick={() => (showForm = !showForm)}>
			{showForm ? 'Cancel' : '+ New niveau'}
		</Button>
	</header>

	{#if showForm}
		<Card.Root>
			<Card.Content class="pt-6">
				<form method="POST" use:enhance action="?/create" class="space-y-4">
					<div class="grid grid-cols-3 gap-4">
						<div class="space-y-2">
							<Label for="code">Code *</Label>
							<Input id="code" name="code" required placeholder="CE1" />
						</div>
						<div class="col-span-2 space-y-2">
							<Label for="label">Label *</Label>
							<Input id="label" name="label" required placeholder="Cours Élémentaire 1" />
						</div>
					</div>
					<div class="space-y-2">
						<Label for="sort_order">Sort order</Label>
						<Input id="sort_order" name="sort_order" type="number" min="0" value="0" />
						<p class="text-xs text-muted-foreground">Lower values appear first. Use 10/20/30 to leave room for inserting later.</p>
					</div>
					{#if form?.error}
						<Alert variant="destructive"><AlertDescription>{form.error}</AlertDescription></Alert>
					{/if}
					<div class="flex justify-end">
						<Button type="submit" size="sm">Create</Button>
					</div>
				</form>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if data.niveaux.length === 0}
		<Card.Root>
			<Card.Content class="py-12 text-center text-sm text-muted-foreground">
				No niveaux yet.
			</Card.Content>
		</Card.Root>
	{:else}
		<Card.Root>
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Order</Table.Head>
						<Table.Head>Code</Table.Head>
						<Table.Head>Label</Table.Head>
						<Table.Head class="w-px"></Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.niveaux as n}
						<Table.Row>
							<Table.Cell class="text-muted-foreground">{n.sort_order}</Table.Cell>
							<Table.Cell class="font-medium">{n.code}</Table.Cell>
							<Table.Cell>{n.label}</Table.Cell>
							<Table.Cell>
								<form method="POST" use:enhance action="?/delete">
									<input type="hidden" name="id" value={n.id} />
									<Button
										type="submit"
										variant="ghost"
										size="sm"
										class="h-7 px-2 text-xs text-destructive hover:text-destructive"
										onclick={(e) => { if (!confirm('Delete this niveau?')) e.preventDefault(); }}
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
	{/if}
</div>
