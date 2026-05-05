<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { enhance } from '$app/forms';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Badge } from '$lib/components/ui/badge';
	import FormToaster from '$lib/components/form-toaster.svelte';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	function statusVariant(s: string) {
		switch (s) {
			case 'done': return 'default';
			case 'pending':
			case 'running': return 'secondary';
			case 'failed': return 'destructive';
			default: return 'secondary';
		}
	}
</script>

<div class="space-y-6">
	<FormToaster {form} />

	<header>
		<h1 class="text-2xl font-semibold">AI</h1>
		<p class="mt-1 text-sm text-muted-foreground">
			LLM-backed item generation, rubric drafting, free-text grading, and distractor generation.
		</p>
	</header>

	<Card.Root>
		<Card.Header><Card.Title class="text-base">Provider</Card.Title></Card.Header>
		<Card.Content>
			{#if data.provider}
				<dl class="grid grid-cols-2 gap-y-2 text-sm">
					<dt class="text-muted-foreground">Provider</dt>
					<dd>
						{#if data.provider.provider === 'openai'}<Badge>OpenAI (live)</Badge>
						{:else if data.provider.provider === 'stub'}<Badge variant="secondary">Stub (dev)</Badge>
							<span class="ml-2 text-xs text-muted-foreground">No <code>OPENAI_API_KEY</code> set; canned responses only.</span>
						{:else}<Badge variant="destructive">{data.provider.provider}</Badge>{/if}
					</dd>
					<dt class="text-muted-foreground">Model</dt>
					<dd class="font-mono text-xs">{data.provider.model}</dd>
				</dl>
			{:else}
				<p class="text-sm text-muted-foreground">Provider info unavailable.</p>
			{/if}
		</Card.Content>
	</Card.Root>

	<div class="grid gap-4 lg:grid-cols-2">
		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">Suggest assessment items</Card.Title>
				<Card.Description>Generate questions for a topic at a given scale + niveau.</Card.Description>
			</Card.Header>
			<Card.Content>
				<form method="POST" action="?/suggest" use:enhance class="space-y-3">
					<div class="space-y-1.5">
						<Label for="topic">Topic *</Label>
						<Input id="topic" name="topic" required placeholder="e.g. addition with carrying" />
					</div>
					<div class="grid grid-cols-3 gap-2">
						<div class="space-y-1.5">
							<Label for="scale_code" class="text-xs">Scale</Label>
							<select id="scale_code" name="scale_code" class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs">
								<option value="french_5level">French</option>
								<option value="maths_5level">Maths</option>
							</select>
						</div>
						<div class="space-y-1.5">
							<Label for="niveau_label" class="text-xs">Niveau</Label>
							<Input id="niveau_label" name="niveau_label" value="CE1" />
						</div>
						<div class="space-y-1.5">
							<Label for="count" class="text-xs">Count</Label>
							<Input id="count" name="count" type="number" min="1" max="10" value="3" />
						</div>
					</div>
					<div class="flex justify-end">
						<Button type="submit" size="sm">Generate</Button>
					</div>
				</form>

				{#if form?.suggested}
					<div class="mt-4 space-y-2">
						<p class="text-xs font-semibold uppercase text-muted-foreground">Suggested items</p>
						{#each form.suggested as item}
							<div class="rounded-md border p-3">
								<div class="flex items-center gap-2">
									<Badge variant="secondary">{item.type}</Badge>
									{#if item.required}<Badge>Required</Badge>{/if}
								</div>
								<p class="mt-1 text-sm font-medium">{item.title}</p>
								{#if item.options && item.options.length > 0}
									<ul class="mt-2 space-y-0.5 pl-5 text-xs text-muted-foreground">
										{#each item.options as opt}<li class="list-disc">{opt}</li>{/each}
									</ul>
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">Generate MCQ distractors</Card.Title>
				<Card.Description>Plausible-but-wrong answers for a multiple-choice question.</Card.Description>
			</Card.Header>
			<Card.Content>
				<form method="POST" action="?/distractors" use:enhance class="space-y-3">
					<div class="space-y-1.5">
						<Label for="question_title">Question *</Label>
						<Input id="question_title" name="question_title" required placeholder="What is the capital of Guinea?" />
					</div>
					<div class="space-y-1.5">
						<Label for="correct_answer">Correct answer *</Label>
						<Input id="correct_answer" name="correct_answer" required placeholder="Conakry" />
					</div>
					<div class="space-y-1.5">
						<Label for="dist_count" class="text-xs">Count</Label>
						<Input id="dist_count" name="count" type="number" min="1" max="10" value="3" />
					</div>
					<div class="flex justify-end">
						<Button type="submit" size="sm">Generate</Button>
					</div>
				</form>

				{#if form?.distractors}
					<div class="mt-4">
						<p class="text-xs font-semibold uppercase text-muted-foreground">Distractors</p>
						<ul class="mt-2 space-y-1">
							{#each form.distractors as d}
								<li class="rounded-md border bg-muted/30 px-3 py-1.5 text-sm">{d}</li>
							{/each}
						</ul>
					</div>
				{/if}
			</Card.Content>
		</Card.Root>
	</div>

	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">Recent jobs</Card.Title>
			<Card.Description>Audit log of every LLM call (kind, model, latency, tokens).</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if data.jobs.length === 0}
				<p class="py-8 text-center text-sm text-muted-foreground">No AI jobs yet.</p>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>When</Table.Head>
							<Table.Head>Kind</Table.Head>
							<Table.Head>Model</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head>Tokens</Table.Head>
							<Table.Head>Latency</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each data.jobs as j}
							<Table.Row>
								<Table.Cell class="text-xs text-muted-foreground">{new Date(j.created_at).toLocaleString()}</Table.Cell>
								<Table.Cell><Badge variant="secondary">{j.kind}</Badge></Table.Cell>
								<Table.Cell class="font-mono text-xs">{j.model}</Table.Cell>
								<Table.Cell><Badge variant={statusVariant(j.status)}>{j.status}</Badge></Table.Cell>
								<Table.Cell class="text-muted-foreground">{j.request_tokens + j.response_tokens}</Table.Cell>
								<Table.Cell class="text-muted-foreground">{j.latency_ms}ms</Table.Cell>
							</Table.Row>
							{#if j.error}
								<Table.Row>
									<Table.Cell></Table.Cell>
									<Table.Cell colspan={5} class="pt-0 text-xs text-destructive">⚠ {j.error}</Table.Cell>
								</Table.Row>
							{/if}
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
