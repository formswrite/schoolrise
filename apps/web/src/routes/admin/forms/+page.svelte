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
	import { enhance } from '$app/forms';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	let showForm = $state(false);
</script>

<div class="space-y-6">
	<FormToaster {form} />
	<header class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold">Forms</h1>
			<p class="mt-1 text-sm text-muted-foreground">
				Build assessment questionnaires. Publish to lock a version, then attach to a campaign.
			</p>
		</div>
		<Button size="sm" onclick={() => (showForm = !showForm)}>
			{showForm ? 'Cancel' : '+ New form'}
		</Button>
	</header>

	{#if showForm}
		<Card.Root>
			<Card.Content class="pt-6">
				<form method="POST" use:enhance action="?/create" class="space-y-4">
					<div class="space-y-2">
						<Label for="title">Title *</Label>
						<Input id="title" name="title" required placeholder="French Q1 2025-2026" />
					</div>
					<div class="space-y-2">
						<Label for="description">Description</Label>
						<Input id="description" name="description" placeholder="optional" />
					</div>
					{#if form?.error}
						<Alert variant="destructive"><AlertDescription>{form.error}</AlertDescription></Alert>
					{/if}
					<div class="flex justify-end">
						<Button type="submit" size="sm">Create form</Button>
					</div>
				</form>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if data.forms.length === 0}
		<Card.Root>
			<Card.Content class="py-12 text-center text-sm text-muted-foreground">
				No forms yet.
			</Card.Content>
		</Card.Root>
	{:else}
		<Card.Root>
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Title</Table.Head>
						<Table.Head>Status</Table.Head>
						<Table.Head>Public ID</Table.Head>
						<Table.Head>Updated</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.forms as f}
						<Table.Row>
							<Table.Cell class="font-medium">
								<a href="/admin/forms/{f.id}" class="text-[#6439B5] hover:underline">{f.title}</a>
							</Table.Cell>
							<Table.Cell>
								{#if f.status === 'published'}<Badge>Published</Badge>
								{:else if f.status === 'closed'}<Badge variant="destructive">Closed</Badge>
								{:else}<Badge variant="secondary">Draft</Badge>{/if}
							</Table.Cell>
							<Table.Cell class="font-mono text-xs">{f.public_id}</Table.Cell>
							<Table.Cell class="text-muted-foreground"
								>{new Date(f.updated_at).toLocaleString()}</Table.Cell
							>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</Card.Root>
	{/if}
</div>
