<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { enhance } from '$app/forms';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Badge } from '$lib/components/ui/badge';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import FormToaster from '$lib/components/form-toaster.svelte';
	import ScopePicker from '$lib/components/scope-picker.svelte';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	let showForm = $state(false);

	const publishedForms = $derived(data.forms.filter((f) => f.status === 'published'));
	let selectedFormID = $state<number | null>(null);
</script>

<div class="space-y-6">
	<FormToaster {form} />

	<header class="space-y-3">
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-semibold">Campaigns</h1>
				{#if data.scope}
					<p class="mt-1 text-sm text-muted-foreground">
						at <a
							href="/admin/institutions?parent={data.scope.id}"
							class="font-medium text-foreground hover:underline">{data.scope.label}</a
						>
					</p>
				{/if}
			</div>
			{#if data.scopeNodeID}
				<Button size="sm" onclick={() => (showForm = !showForm)}>
					{showForm ? 'Cancel' : '+ New campaign'}
				</Button>
			{/if}
		</div>
		{#if data.scopeOptions && data.scopeOptions.length > 0 && data.scopeNodeID}
			<ScopePicker options={data.scopeOptions} current={data.scopeNodeID} />
		{/if}
	</header>

	{#if !data.scopeNodeID}
		<Alert>
			<AlertDescription>
				No scopes are configured yet. Set up your hierarchy first via <a
					href="/admin/institutions"
					class="font-medium underline">Institutions</a
				>.
			</AlertDescription>
		</Alert>
	{:else}
		{#if showForm}
			<Card.Root>
				<Card.Header>
					<Card.Title class="text-base">New campaign</Card.Title>
					<Card.Description
						>Bind a published form version to a scale + period at this scope.</Card.Description
					>
				</Card.Header>
				<Card.Content>
					{#if publishedForms.length === 0 || data.periods.length === 0}
						<Alert variant="destructive">
							<AlertDescription>
								Need at least one <strong>published</strong> form ({publishedForms.length}) and one
								<strong>period</strong>
								({data.periods.length}). Create them in
								<a href="/admin/forms" class="underline">Forms</a>
								and <a href="/admin/periods" class="underline">Periods</a>.
							</AlertDescription>
						</Alert>
					{:else}
						<form method="POST" action="?/create" use:enhance class="space-y-4">
							<div class="space-y-2">
								<Label for="title">Title *</Label>
								<Input id="title" name="title" required placeholder="French Q1 2025-2026" />
							</div>
							<div class="grid grid-cols-2 gap-4">
								<div class="space-y-2">
									<Label for="scale_code">Scale *</Label>
									<select
										id="scale_code"
										name="scale_code"
										required
										class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus:ring-1 focus:ring-ring focus:outline-none"
									>
										{#each data.scales as s}<option value={s.code}>{s.label}</option>{/each}
									</select>
								</div>
								<div class="space-y-2">
									<Label for="period_id">Period *</Label>
									<select
										id="period_id"
										name="period_id"
										required
										class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus:ring-1 focus:ring-ring focus:outline-none"
									>
										{#each data.periods as p}
											<option value={p.id} selected={p.is_current}
												>{p.label}{p.is_current ? ' (current)' : ''}</option
											>
										{/each}
									</select>
								</div>
							</div>
							<div class="grid grid-cols-2 gap-4">
								<div class="space-y-2">
									<Label for="form_id">Form *</Label>
									<select
										id="form_id"
										name="form_id"
										required
										bind:value={selectedFormID}
										class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus:ring-1 focus:ring-ring focus:outline-none"
									>
										<option value={null}></option>
										{#each publishedForms as f}<option value={f.id}>{f.title}</option>{/each}
									</select>
								</div>
								<div class="space-y-2">
									<Label for="form_version_id">Form version ID *</Label>
									<Input
										id="form_version_id"
										name="form_version_id"
										type="number"
										min="1"
										required
										placeholder="e.g. 1"
									/>
									<p class="text-xs text-muted-foreground">
										Find via <a href="/admin/forms/{selectedFormID ?? ''}" class="underline"
											>Forms</a
										> → versions list.
									</p>
								</div>
							</div>
							<div class="flex justify-end">
								<Button type="submit" size="sm">Create campaign</Button>
							</div>
						</form>
					{/if}
				</Card.Content>
			</Card.Root>
		{/if}

		{#if data.campaigns.length === 0}
			<Card.Root>
				<Card.Content class="py-12 text-center text-sm text-muted-foreground">
					No campaigns at this scope yet.
				</Card.Content>
			</Card.Root>
		{:else}
			<Card.Root>
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Campaign</Table.Head>
							<Table.Head>Scale</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head>Created</Table.Head>
							<Table.Head class="w-px"></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each data.campaigns as c}
							<Table.Row>
								<Table.Cell class="font-medium">
									<a href="/admin/campaigns/{c.id}" class="text-[#6439B5] hover:underline"
										>{c.title}</a
									>
								</Table.Cell>
								<Table.Cell><Badge variant="secondary">{c.scale_code}</Badge></Table.Cell>
								<Table.Cell>
									{#if c.status === 'open'}<Badge>Open</Badge>
									{:else if c.status === 'closed'}<Badge variant="destructive">Closed</Badge>
									{:else}<Badge variant="secondary">Draft</Badge>{/if}
								</Table.Cell>
								<Table.Cell class="text-muted-foreground"
									>{new Date(c.created_at).toLocaleDateString()}</Table.Cell
								>
								<Table.Cell>
									<Button
										href="/admin/campaigns/{c.id}"
										variant="ghost"
										size="sm"
										class="h-7 text-xs">Open →</Button
									>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Root>
		{/if}
	{/if}
</div>
