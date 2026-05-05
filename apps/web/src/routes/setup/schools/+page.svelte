<script lang="ts">
	import type { ActionData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Alert, AlertDescription, AlertTitle } from '$lib/components/ui/alert';

	let { form }: { form: ActionData } = $props();
</script>

<Card.Root>
	<Card.Header>
		<Card.Title class="text-2xl">Bulk import schools (optional)</Card.Title>
		<Card.Description>
			Upload a CSV with columns <code>parent_code,level,code,label</code>. Parents must appear
			before children.
		</Card.Description>
	</Card.Header>
	<Card.Content>
		<details class="mb-4 text-sm">
			<summary class="cursor-pointer text-muted-foreground hover:text-foreground"
				>Show example</summary
			>
			<pre class="mt-2 overflow-x-auto rounded-md bg-muted p-3 text-xs">parent_code,level,code,label
,region,N,North Region
,region,S,South Region
N,school,N-001,Roosevelt High
N,school,N-002,Lincoln Elementary
S,school,S-001,Adams Middle</pre>
		</details>

		<form method="POST" action="?/import" class="space-y-4">
			<Textarea
				name="csv"
				rows={10}
				required
				placeholder="parent_code,level,code,label..."
				class="font-mono text-sm"
			/>

			{#if form?.error}
				<Alert variant="destructive">
					<AlertDescription>{form.error}</AlertDescription>
				</Alert>
			{/if}

			{#if form?.imported !== undefined}
				<Alert>
					<AlertTitle>Imported {form.imported} nodes.</AlertTitle>
				</Alert>
			{/if}

			{#if form?.errors && form.errors.length > 0}
				<Alert variant="destructive">
					<AlertTitle>Some rows failed:</AlertTitle>
					<AlertDescription>
						<ul class="ml-5 list-disc text-xs">
							{#each form.errors as err}
								<li>{err}</li>
							{/each}
						</ul>
					</AlertDescription>
				</Alert>
			{/if}

			<div class="flex justify-between">
				<Button variant="ghost" href="/setup/levels">Back</Button>
				<div class="flex gap-2">
					<Button type="submit" formaction="?/skip" variant="outline">Skip</Button>
					<Button type="submit">Import</Button>
				</div>
			</div>
		</form>
	</Card.Content>
</Card.Root>
