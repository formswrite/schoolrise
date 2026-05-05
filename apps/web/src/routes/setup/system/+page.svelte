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
		<Card.Title class="text-2xl">System settings</Card.Title>
		<Card.Description>You can change these later in admin settings.</Card.Description>
	</Card.Header>
	<Card.Content>
		<form method="POST" class="space-y-4">
			<div class="space-y-2">
				<Label for="instance_name">Instance name</Label>
				<Input
					id="instance_name"
					name="instance_name"
					type="text"
					required
					value={form?.instanceName ?? ''}
					placeholder="Ministry of Education — Country"
				/>
			</div>
			<div class="grid grid-cols-2 gap-4">
				<div class="space-y-2">
					<Label for="default_locale">Default locale</Label>
					<select
						id="default_locale"
						name="default_locale"
						class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs transition-colors focus:outline-none focus:ring-1 focus:ring-ring"
					>
						<option value="en" selected={form?.defaultLocale !== 'fr'}>English</option>
						<option value="fr" selected={form?.defaultLocale === 'fr'}>Français</option>
					</select>
				</div>
				<div class="space-y-2">
					<Label for="time_zone">Time zone</Label>
					<Input id="time_zone" name="time_zone" type="text" value={form?.timeZone ?? 'UTC'} placeholder="UTC, Africa/Conakry…" />
				</div>
			</div>
			<div class="space-y-2">
				<Label for="base_url">Public base URL</Label>
				<Input id="base_url" name="base_url" type="url" required value={form?.baseURL ?? ''} placeholder="https://schoolrise.minedu.gov.example" />
				<p class="text-xs text-muted-foreground">Used in assignment-link emails and API responses.</p>
			</div>
			{#if form?.error}
				<Alert variant="destructive">
					<AlertDescription>{form.error}</AlertDescription>
				</Alert>
			{/if}
			<div class="flex justify-between">
				<Button variant="ghost" href="/setup/admin">Back</Button>
				<Button type="submit">Continue</Button>
			</div>
		</form>
	</Card.Content>
</Card.Root>
