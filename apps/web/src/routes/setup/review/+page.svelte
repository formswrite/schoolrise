<script lang="ts">
	import type { ActionData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';

	let { form }: { form: ActionData } = $props();
</script>

<Card.Root>
	<Card.Header>
		<Card.Title class="text-2xl">Ready to finalize</Card.Title>
		<Card.Description>
			After finalize the setup endpoints can no longer be called — re-running setup requires a database reset.
		</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-4">
		<ul class="space-y-2 text-sm">
			<li>✓ Administrator account created</li>
			<li>✓ System settings saved</li>
			<li>✓ Hierarchy levels configured</li>
			<li>✓ Schools imported (or skipped)</li>
			<li>✓ Integrations configured (or skipped)</li>
			<li>✓ SMTP relay configured (or skipped)</li>
		</ul>
		<form method="POST">
			{#if form?.error}
				<Alert variant="destructive" class="mb-4">
					<AlertDescription>{form.error}</AlertDescription>
				</Alert>
			{/if}
			<div class="flex justify-between">
				<Button variant="ghost" href="/setup/smtp">Back</Button>
				<Button type="submit" class="bg-emerald-600 text-white hover:bg-emerald-700">
					Finalize installation
				</Button>
			</div>
		</form>
	</Card.Content>
</Card.Root>
