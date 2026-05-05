<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Card from '$lib/components/ui/card';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<div class="flex min-h-screen items-center justify-center bg-muted p-4">
	<Card.Root class="w-full max-w-sm">
		<Card.Header>
			<Card.Title class="text-2xl">Change your password</Card.Title>
			{#if data.user.mustChangePassword}
				<Card.Description class="text-amber-700">
					Your administrator requires you to change your password before continuing.
				</Card.Description>
			{:else}
				<Card.Description>Choose a new password to use going forward.</Card.Description>
			{/if}
		</Card.Header>
		<Card.Content>
			<form method="POST" class="space-y-4">
				<div class="space-y-2">
					<Label for="current_password">Current password</Label>
					<Input
						id="current_password"
						type="password"
						name="current_password"
						required
						autocomplete="current-password"
					/>
				</div>

				<div class="space-y-2">
					<Label for="new_password">New password</Label>
					<Input
						id="new_password"
						type="password"
						name="new_password"
						required
						minlength={8}
						autocomplete="new-password"
					/>
				</div>

				<div class="space-y-2">
					<Label for="confirm_password">Confirm new password</Label>
					<Input
						id="confirm_password"
						type="password"
						name="confirm_password"
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

				<Button type="submit" class="w-full">Change password</Button>
			</form>
		</Card.Content>
	</Card.Root>
</div>
