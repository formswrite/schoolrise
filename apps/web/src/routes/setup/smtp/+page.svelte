<script lang="ts">
	import type { ActionData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';

	let { form }: { form: ActionData } = $props();

	let useTLS = $state(true);
</script>

<Card.Root>
	<Card.Header>
		<Card.Title class="text-2xl">SMTP relay (optional)</Card.Title>
		<Card.Description>Use an internal mail relay alongside or instead of Resend.</Card.Description>
	</Card.Header>
	<Card.Content>
		<form method="POST" action="?/save" class="space-y-4">
			<div class="grid grid-cols-3 gap-3">
				<div class="col-span-2 space-y-2">
					<Label for="host">SMTP host</Label>
					<Input id="host" name="host" required placeholder="smtp.example.gov" />
				</div>
				<div class="space-y-2">
					<Label for="port">Port</Label>
					<Input id="port" name="port" type="number" value="587" required />
				</div>
			</div>
			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-2">
					<Label for="username">Username</Label>
					<Input id="username" name="username" />
				</div>
				<div class="space-y-2">
					<Label for="password">Password</Label>
					<Input id="password" name="password" type="password" />
				</div>
			</div>
			<div class="space-y-2">
				<Label for="from_address">From address</Label>
				<Input id="from_address" name="from_address" type="email" required placeholder="schoolrise@example.gov" />
			</div>
			<div class="flex items-center gap-2">
				<Checkbox id="use_tls" name="use_tls" bind:checked={useTLS} />
				<Label for="use_tls">Use TLS</Label>
			</div>

			{#if form?.error}
				<Alert variant="destructive">
					<AlertDescription>{form.error}</AlertDescription>
				</Alert>
			{/if}

			<div class="flex justify-between">
				<Button variant="ghost" href="/setup/integrations">Back</Button>
				<div class="flex gap-2">
					<Button type="submit" formaction="?/skip" variant="outline">Skip</Button>
					<Button type="submit">Save</Button>
				</div>
			</div>
		</form>
	</Card.Content>
</Card.Root>
