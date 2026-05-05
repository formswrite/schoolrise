<script lang="ts">
	import type { ActionData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';

	let { form }: { form: ActionData } = $props();

	let mustChange = $state(true);
</script>

<div class="space-y-6">
	<header>
		<a href="/admin/users" class="text-xs text-muted-foreground hover:underline">← Users</a>
		<h1 class="mt-1 text-2xl font-semibold">New user</h1>
		<p class="text-sm text-muted-foreground">Create a user account. Assign roles after creation from the user detail page.</p>
	</header>

	<Card.Root>
		<Card.Content class="pt-6">
			<form method="POST" class="space-y-4">
				<div class="space-y-2">
					<Label for="full_name">Full name</Label>
					<Input id="full_name" name="full_name" required value={form?.fullName ?? ''} />
				</div>
				<div class="space-y-2">
					<Label for="email">Email</Label>
					<Input id="email" name="email" type="email" required autocomplete="off" value={form?.email ?? ''} />
				</div>
				<div class="space-y-2">
					<Label for="role">Role</Label>
					<select
						id="role"
						name="role"
						class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus:outline-none focus:ring-1 focus:ring-ring"
					>
						<option value="teacher" selected={form?.role !== 'admin' && form?.role !== 'inspector'}>Teacher</option>
						<option value="inspector" selected={form?.role === 'inspector'}>Inspector</option>
						<option value="admin" selected={form?.role === 'admin'}>Admin</option>
					</select>
					<p class="text-xs text-muted-foreground">Role label is informational. Effective access is granted via assignments on the next screen.</p>
				</div>
				<div class="space-y-2">
					<Label for="password">Initial password</Label>
					<Input id="password" name="password" type="password" required minlength={8} autocomplete="new-password" />
				</div>
				<div class="flex items-center gap-2">
					<Checkbox id="must_change_password" name="must_change_password" bind:checked={mustChange} />
					<Label for="must_change_password">Force password change on first login</Label>
				</div>

				{#if form?.error}
					<Alert variant="destructive">
						<AlertDescription>{form.error}</AlertDescription>
					</Alert>
				{/if}

				<div class="flex justify-between">
					<Button variant="ghost" href="/admin/users">Cancel</Button>
					<Button type="submit">Create user</Button>
				</div>
			</form>
		</Card.Content>
	</Card.Root>
</div>
