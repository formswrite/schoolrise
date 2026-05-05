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

	let { form }: { data: PageData; form: ActionData } = $props();

	const sampleCSV = `full_name,gender,date_of_birth,student_code,enrollment_date
Alice Diallo,female,2010-04-12,STU-A-001,2025-09-01
Bob Camara,male,2009-11-30,STU-A-002,2025-09-01
Cleo Bah,female,2010-02-28,STU-A-003,2025-09-01`;
</script>

<div class="space-y-6">
	<FormToaster {form} />
	<header>
		<h1 class="text-2xl font-semibold">CSV import — students</h1>
		<p class="mt-1 text-sm text-muted-foreground">
			Upload a CSV (or paste it) to bulk-create students. Use <strong>dry-run</strong> first to preview.
		</p>
	</header>

	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">Required headers</Card.Title>
			<Card.Description>
				<code>full_name</code> is required.
				Optional: <code>given_name</code>, <code>family_name</code>, <code>gender</code>,
				<code>date_of_birth</code> (YYYY-MM-DD), <code>email</code>, <code>phone</code>,
				<code>student_code</code>, <code>enrollment_date</code> (YYYY-MM-DD).
			</Card.Description>
		</Card.Header>
		<Card.Content>
			<details>
				<summary class="cursor-pointer text-sm font-medium text-[#6439B5]">Sample CSV</summary>
				<pre class="mt-2 overflow-x-auto rounded-md border bg-muted/40 p-3 text-xs">{sampleCSV}</pre>
			</details>
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Content class="pt-6">
			<form method="POST" use:enhance action="?/upload" enctype="multipart/form-data" class="space-y-4">
				<div class="grid grid-cols-2 gap-4">
					<div class="space-y-2">
						<Label for="institution_id">Institution ID *</Label>
						<Input id="institution_id" name="institution_id" type="number" min="1" required placeholder="e.g. 12" />
						<p class="text-xs text-muted-foreground">
							Find the ID via <a href="/admin/institutions" class="underline">/admin/institutions</a> (URL has <code>?parent=NN</code>).
						</p>
					</div>
					<div class="flex items-end">
						<div class="flex items-center gap-2">
							<Checkbox id="dry_run" name="dry_run" checked />
							<Label for="dry_run" class="cursor-pointer">Dry-run (preview only, no inserts)</Label>
						</div>
					</div>
				</div>

				<div class="space-y-2">
					<Label for="csv_file">CSV file</Label>
					<Input id="csv_file" name="csv_file" type="file" accept=".csv,text/csv" />
				</div>

				<div class="space-y-2">
					<Label for="csv_pasted">…or paste CSV text</Label>
					<textarea
						id="csv_pasted"
						name="csv_pasted"
						rows="6"
						placeholder="paste here if you don't want to upload a file"
						class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 font-mono text-sm shadow-xs focus:outline-none focus:ring-1 focus:ring-ring"
					></textarea>
				</div>

				{#if form?.error}
					<Alert variant="destructive"><AlertDescription>{form.error}</AlertDescription></Alert>
				{/if}

				<div class="flex justify-end">
					<Button type="submit">Run import</Button>
				</div>
			</form>
		</Card.Content>
	</Card.Root>

	{#if form?.job}
		{@const job = form.job}
		<Card.Root>
			<Card.Header>
				<div class="flex items-center justify-between">
					<Card.Title class="text-base">
						Import #{job.id}
						{#if job.dry_run}<Badge variant="secondary" class="ml-2">Dry run</Badge>{/if}
					</Card.Title>
					{#if job.status === 'completed'}<Badge>Completed</Badge>
					{:else if job.status === 'failed'}<Badge variant="destructive">Failed</Badge>
					{:else}<Badge variant="secondary">{job.status}</Badge>{/if}
				</div>
				<Card.Description>
					{job.total_rows} rows · {job.succeeded} succeeded · {job.failed} failed
				</Card.Description>
			</Card.Header>
			{#if job.errors.length > 0}
				<Card.Content>
					<h4 class="mb-2 text-sm font-semibold">Row errors</h4>
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Row #</Table.Head>
								<Table.Head>Field</Table.Head>
								<Table.Head>Error</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each job.errors as e}
								<Table.Row>
									<Table.Cell class="font-mono text-xs">{e.row_number}</Table.Cell>
									<Table.Cell class="text-muted-foreground">{e.field || '—'}</Table.Cell>
									<Table.Cell class="text-destructive">{e.error}</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			{/if}
		</Card.Root>
	{/if}
</div>
