<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Badge } from '$lib/components/ui/badge';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import FormToaster from '$lib/components/form-toaster.svelte';
	import ScopePicker from '$lib/components/scope-picker.svelte';
	import { enhance } from '$app/forms';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let showForm = $state(false);

	const niveauLabel = (id: number) => data.niveaux.find((n) => n.id === id)?.label ?? `#${id}`;
	const periodLabel = (id: number) => {
		const p = data.periods.find((p) => p.id === id);
		return p ? `${p.label}${p.is_current ? ' (current)' : ''}` : `#${id}`;
	};
</script>

<div class="space-y-6">
	<FormToaster {form} />
	<header class="space-y-3">
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-semibold">Classes</h1>
				{#if data.institution}
					<p class="mt-1 text-sm text-muted-foreground">
						at <a href="/admin/institutions?parent={data.institution.id}" class="font-medium text-foreground hover:underline">{data.institution.label}</a>
					</p>
				{/if}
			</div>
			{#if data.institutionId}
				<Button size="sm" onclick={() => (showForm = !showForm)}>
					{showForm ? 'Cancel' : '+ New class'}
				</Button>
			{/if}
		</div>
		{#if data.institutionOptions && data.institutionOptions.length > 0 && data.institutionId}
			<ScopePicker options={data.institutionOptions} current={data.institutionId} paramName="institution" />
		{/if}
	</header>

	{#if !data.institutionId}
		<Alert>
			<AlertDescription>
				No institutions exist yet. Create one first via <a href="/admin/institutions" class="font-medium underline">Institutions</a>.
			</AlertDescription>
		</Alert>
	{:else}
		{#if data.periods.length === 0 || data.niveaux.length === 0}
			<Alert variant="destructive">
				<AlertDescription>
					You need at least one academic period and one niveau before creating a class.
					Currently: {data.periods.length} period(s), {data.niveaux.length} niveau(x).
				</AlertDescription>
			</Alert>
		{/if}

		{#if showForm && data.periods.length > 0 && data.niveaux.length > 0}
			<Card.Root>
				<Card.Header>
					<Card.Title class="text-base">New class</Card.Title>
				</Card.Header>
				<Card.Content>
					<form method="POST" use:enhance action="?/create" class="space-y-4">
						<div class="grid grid-cols-2 gap-4">
							<div class="space-y-2">
								<Label for="period_id">Period *</Label>
								<select
									id="period_id"
									name="period_id"
									required
									class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus:outline-none focus:ring-1 focus:ring-ring"
								>
									{#each data.periods as p}
										<option value={p.id} selected={p.is_current}>{p.label}{p.is_current ? ' (current)' : ''}</option>
									{/each}
								</select>
							</div>
							<div class="space-y-2">
								<Label for="niveau_id">Niveau *</Label>
								<select
									id="niveau_id"
									name="niveau_id"
									required
									class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus:outline-none focus:ring-1 focus:ring-ring"
								>
									{#each data.niveaux as n}
										<option value={n.id}>{n.label} ({n.code})</option>
									{/each}
								</select>
							</div>
						</div>
						<div class="grid grid-cols-3 gap-4">
							<div class="space-y-2">
								<Label for="code">Code *</Label>
								<Input id="code" name="code" required placeholder="CE1-A" />
							</div>
							<div class="col-span-2 space-y-2">
								<Label for="label">Label *</Label>
								<Input id="label" name="label" required placeholder="CE1 Group A — morning" />
							</div>
						</div>
						<div class="space-y-2">
							<Label for="capacity">Capacity (optional)</Label>
							<Input id="capacity" name="capacity" type="number" min="0" placeholder="30" />
						</div>

						{#if form?.error}
							<Alert variant="destructive">
								<AlertDescription>{form.error}</AlertDescription>
							</Alert>
						{/if}

						<div class="flex justify-end">
							<Button type="submit" size="sm">Create class</Button>
						</div>
					</form>
				</Card.Content>
			</Card.Root>
		{/if}

		{#if data.classes.length === 0}
			<Card.Root>
				<Card.Content class="py-12 text-center text-sm text-muted-foreground">
					No classes here yet.
				</Card.Content>
			</Card.Root>
		{:else}
			<Card.Root>
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Code</Table.Head>
							<Table.Head>Label</Table.Head>
							<Table.Head>Niveau</Table.Head>
							<Table.Head>Period</Table.Head>
							<Table.Head>Capacity</Table.Head>
							<Table.Head class="w-px"></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each data.classes as c}
							<Table.Row>
								<Table.Cell class="font-medium">
									<a href="/admin/classes/{c.id}" class="text-[#6439B5] hover:underline">{c.code}</a>
								</Table.Cell>
								<Table.Cell>{c.label}</Table.Cell>
								<Table.Cell><Badge variant="secondary">{niveauLabel(c.niveau_id)}</Badge></Table.Cell>
								<Table.Cell class="text-muted-foreground">{periodLabel(c.period_id)}</Table.Cell>
								<Table.Cell class="text-muted-foreground">{c.capacity || '—'}</Table.Cell>
								<Table.Cell>
									<form method="POST" use:enhance action="?/delete">
										<input type="hidden" name="id" value={c.id} />
										<Button
											type="submit"
											variant="ghost"
											size="sm"
											class="h-7 px-2 text-xs text-destructive hover:text-destructive"
											onclick={(e) => { if (!confirm('Delete this class?')) e.preventDefault(); }}
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
	{/if}
</div>
