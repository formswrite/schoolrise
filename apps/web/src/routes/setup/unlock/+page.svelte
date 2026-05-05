<script lang="ts">
	import type { ActionData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';

	let { form }: { form: ActionData } = $props();
</script>

<Card.Root>
	<Card.Header>
		<Card.Title class="text-2xl">Paste your install token</Card.Title>
		<Card.Description>
			The install token was printed to your container logs on first boot. Run
			<code class="rounded bg-muted px-1.5 py-0.5 text-xs">docker compose logs app | grep token</code>
			to retrieve it.
		</Card.Description>
	</Card.Header>
	<Card.Content>
		<form method="POST" class="space-y-4">
			<div class="space-y-2">
				<Label for="install_token">Install token</Label>
				<Input id="install_token" name="install_token" type="text" required autocomplete="off" class="font-mono" />
			</div>
			{#if form?.error}
				<Alert variant="destructive">
					<AlertDescription>{form.error}</AlertDescription>
				</Alert>
			{/if}
			<div class="flex justify-between">
				<Button variant="ghost" href="/setup/welcome">Back</Button>
				<Button type="submit">Unlock</Button>
			</div>
		</form>
	</Card.Content>
</Card.Root>
