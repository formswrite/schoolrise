<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { Badge } from '$lib/components/ui/badge';
	import FormToaster from '$lib/components/form-toaster.svelte';
	import { enhance } from '$app/forms';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let showAddForm = $state(false);

	function levelLabel(code: string): string {
		return data.levels.find((l) => l.code === code)?.label ?? code;
	}
</script>

<div class="space-y-6">
	<FormToaster {form} />
	<header class="flex items-start justify-between">
		<div>
			<h1 class="text-2xl font-semibold">Institutions</h1>
			<nav class="mt-2 flex flex-wrap items-center gap-1 text-sm text-muted-foreground">
				<a href="/admin/institutions" class="rounded px-1.5 py-0.5 hover:bg-accent">All</a>
				{#each data.breadcrumbs as crumb}
					<span class="text-muted-foreground/50">/</span>
					<a href="/admin/institutions?parent={crumb.id}" class="rounded px-1.5 py-0.5 hover:bg-accent">
						{crumb.label}
						<span class="text-xs text-muted-foreground">({levelLabel(crumb.level)})</span>
					</a>
				{/each}
			</nav>
		</div>
		<div class="flex gap-2">
			{#if data.parentId}
				<Button variant="outline" size="sm" href="/admin/staff?scope={data.parentId}">Staff here</Button>
				{#if data.parent && data.parent.level === 'institution'}
					<Button variant="outline" size="sm" href="/admin/students?institution={data.parentId}">Students</Button>
				{/if}
			{/if}
			{#if data.childLevelCode && !(data.parent && data.parent.level === 'institution')}
				<Button size="sm" onclick={() => (showAddForm = !showAddForm)}>
					{showAddForm ? 'Cancel' : `+ Add ${levelLabel(data.childLevelCode)}`}
				</Button>
			{/if}
		</div>
	</header>

	{#if data.levels.length === 0}
		<Alert>
			<AlertDescription>
				No hierarchy levels are defined. Run the setup wizard or add levels via the database.
			</AlertDescription>
		</Alert>
	{/if}

	{#if showAddForm && data.childLevelCode}
		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">Add {levelLabel(data.childLevelCode)}</Card.Title>
			</Card.Header>
			<Card.Content>
				<form method="POST" use:enhance action="?/create" class="space-y-4">
					<input type="hidden" name="parent_id" value={data.parentId ?? ''} />
					<input type="hidden" name="level" value={data.childLevelCode} />
					<div class="grid grid-cols-2 gap-4">
						<div class="space-y-2">
							<Label for="code">Code</Label>
							<Input id="code" name="code" required placeholder="unique-code" />
						</div>
						<div class="space-y-2">
							<Label for="label">Label</Label>
							<Input id="label" name="label" required placeholder="Display name" />
						</div>
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

	{#if data.nodes.length === 0}
		{#if data.leafSummary && data.parent}
			<Card.Root>
				<Card.Header>
					<Card.Title class="text-base">{data.parent.label}</Card.Title>
					<Card.Description>
						This is a school — the bottom of the hierarchy. Schools contain people and classes, not more institutions.
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<div class="grid gap-3 sm:grid-cols-3">
						<a
							href="/admin/students?institution={data.parentId}"
							class="rounded-md border p-4 transition hover:border-[#6439B5] hover:bg-[#f0ecff]"
						>
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Students</p>
							<p class="mt-1 text-3xl font-bold">{data.leafSummary.studentCount}</p>
							<p class="mt-1 text-xs text-[#6439B5]">View roster →</p>
						</a>
						<a
							href="/admin/staff?scope={data.parentId}"
							class="rounded-md border p-4 transition hover:border-[#6439B5] hover:bg-[#f0ecff]"
						>
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Staff</p>
							<p class="mt-1 text-3xl font-bold">{data.leafSummary.staffCount}</p>
							<p class="mt-1 text-xs text-[#6439B5]">View staff →</p>
						</a>
						<a
							href="/admin/classes?institution={data.parentId}"
							class="rounded-md border p-4 transition hover:border-[#6439B5] hover:bg-[#f0ecff]"
						>
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Classes</p>
							<p class="mt-1 text-3xl font-bold">{data.leafSummary.classCount}</p>
							<p class="mt-1 text-xs text-[#6439B5]">View classes →</p>
						</a>
					</div>
				</Card.Content>
			</Card.Root>
		{:else}
			<Card.Root>
				<Card.Content class="py-12 text-center text-sm text-muted-foreground">
					No nodes here yet. Use <strong>+ Add</strong> to create one.
				</Card.Content>
			</Card.Root>
		{/if}
	{:else}
		<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
			{#each data.nodes as node}
				<Card.Root class="transition hover:shadow-md">
					<Card.Header class="pb-3">
						<a href="/admin/institutions?parent={node.id}" class="block">
							<div class="flex items-baseline justify-between">
								<Card.Title class="text-base">{node.label}</Card.Title>
								<Badge variant="secondary">{levelLabel(node.level)}</Badge>
							</div>
							<Card.Description class="mt-1 text-xs">code: {node.code}</Card.Description>
						</a>
					</Card.Header>
					<Card.Footer class="pt-0">
						<form method="POST" use:enhance action="?/delete">
							<input type="hidden" name="id" value={node.id} />
							<Button
								type="submit"
								variant="ghost"
								size="sm"
								class="h-7 px-2 text-xs text-destructive hover:text-destructive"
								onclick={(e) => { if (!confirm('Delete this node?')) e.preventDefault(); }}
							>
								Delete
							</Button>
						</form>
					</Card.Footer>
				</Card.Root>
			{/each}
		</div>
	{/if}
</div>
