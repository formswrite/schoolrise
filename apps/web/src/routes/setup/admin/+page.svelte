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
		<Card.Title class="text-2xl">Create the administrator account</Card.Title>
		<Card.Description>
			This account will have full access. You will sign in with these credentials after setup.
		</Card.Description>
	</Card.Header>
	<Card.Content>
		<form method="POST" class="space-y-4">
			<div class="space-y-2">
				<Label for="full_name">Full name</Label>
				<Input id="full_name" name="full_name" type="text" required value={form?.fullName ?? ''} />
			</div>
			<div class="space-y-2">
				<Label for="email">Email</Label>
				<Input
					id="email"
					name="email"
					type="email"
					required
					autocomplete="email"
					value={form?.email ?? ''}
				/>
			</div>
			<div class="space-y-2">
				<Label for="password">Password</Label>
				<Input
					id="password"
					name="password"
					type="password"
					required
					minlength={8}
					autocomplete="new-password"
				/>
			</div>
			<div class="space-y-2">
				<Label for="confirm_password">Confirm password</Label>
				<Input
					id="confirm_password"
					name="confirm_password"
					type="password"
					required
					minlength={8}
					autocomplete="new-password"
				/>
			</div>
			{#if form?.error}
				<Alert variant="destructive">
					<AlertDescription>{form.error}</AlertDescription>
				</Alert>
			{/if}
			<div class="flex justify-between">
				<Button variant="ghost" href="/setup/unlock">Back</Button>
				<Button type="submit">Continue</Button>
			</div>
		</form>
	</Card.Content>
</Card.Root>
