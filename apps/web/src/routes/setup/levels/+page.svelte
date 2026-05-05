<script lang="ts">
	import type { ActionData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';

	let { form }: { form: ActionData } = $props();

	type Row = { code: string; label: string; parent: string };

	let rows: Row[] = $state([
		{ code: 'region', label: 'Region', parent: '' },
		{ code: 'school', label: 'School', parent: 'region' }
	]);

	function addRow() {
		const lastCode = rows.length > 0 ? rows[rows.length - 1].code : '';
		rows = [...rows, { code: '', label: '', parent: lastCode }];
	}

	function removeRow(i: number) {
		rows = rows.filter((_, idx) => idx !== i);
	}
</script>

<Card.Root>
	<Card.Header>
		<Card.Title class="text-2xl">Define your hierarchy taxonomy</Card.Title>
		<Card.Description>
			Each level represents a tier. Examples: <code>region → prefecture → school</code> for
			ministries; <code>district → school</code> for districts.
		</Card.Description>
	</Card.Header>
	<Card.Content>
		<form method="POST" class="space-y-4">
			{#each rows as row, i (i)}
				<div class="grid grid-cols-12 gap-2 rounded-md border p-3">
					<div class="col-span-3 space-y-1">
						<Label for="row-code-{i}" class="text-xs">Code</Label>
						<Input
							id="row-code-{i}"
							name="code"
							bind:value={rows[i].code}
							required
							placeholder="region"
						/>
					</div>
					<div class="col-span-4 space-y-1">
						<Label for="row-label-{i}" class="text-xs">Label</Label>
						<Input
							id="row-label-{i}"
							name="label"
							bind:value={rows[i].label}
							required
							placeholder="Region"
						/>
					</div>
					<div class="col-span-4 space-y-1">
						<Label for="row-parent-{i}" class="text-xs">Parent (code)</Label>
						<Input
							id="row-parent-{i}"
							name="parent"
							bind:value={rows[i].parent}
							placeholder={i === 0 ? '(empty for root)' : ''}
						/>
					</div>
					<div class="col-span-1 flex items-end">
						<Button
							type="button"
							variant="ghost"
							size="sm"
							onclick={() => removeRow(i)}
							disabled={rows.length === 1}
							class="text-destructive"
						>
							×
						</Button>
					</div>
				</div>
			{/each}

			<Button type="button" variant="outline" size="sm" onclick={addRow}>+ Add level</Button>

			{#if form?.error}
				<Alert variant="destructive">
					<AlertDescription>{form.error}</AlertDescription>
				</Alert>
			{/if}

			<div class="flex justify-between">
				<Button variant="ghost" href="/setup/system">Back</Button>
				<Button type="submit">Continue</Button>
			</div>
		</form>
	</Card.Content>
</Card.Root>
