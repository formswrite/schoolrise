<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<div class="space-y-6">
	<header>
		<a
			href="/admin/staff?scope={data.scopeNodeId}"
			class="text-xs text-muted-foreground hover:underline">← Staff</a
		>
		<h1 class="mt-1 text-2xl font-semibold">New staff</h1>
		{#if data.node}
			<p class="text-sm text-muted-foreground">
				at <strong>{data.node.label}</strong> ({data.node.level})
			</p>
		{/if}
	</header>

	<Card.Root>
		<Card.Content class="pt-6">
			<form method="POST" class="space-y-4">
				<div class="space-y-2">
					<Label for="full_name">Full name *</Label>
					<Input id="full_name" name="full_name" required />
				</div>
				<div class="grid grid-cols-3 gap-4">
					<div class="col-span-2 space-y-2">
						<Label for="position">Position</Label>
						<Input
							id="position"
							name="position"
							placeholder="e.g. Headteacher, Teacher, Inspector"
						/>
					</div>
					<div class="space-y-2">
						<Label for="staff_code">Staff code</Label>
						<Input id="staff_code" name="staff_code" placeholder="STAFF-001" />
					</div>
				</div>
				<div class="grid grid-cols-3 gap-4">
					<div class="space-y-2">
						<Label for="hire_date">Hire date</Label>
						<Input id="hire_date" name="hire_date" type="date" />
					</div>
					<div class="space-y-2">
						<Label for="email">Email</Label>
						<Input id="email" name="email" type="email" />
					</div>
					<div class="space-y-2">
						<Label for="phone">Phone</Label>
						<Input id="phone" name="phone" type="tel" />
					</div>
				</div>

				{#if form?.error}
					<Alert variant="destructive">
						<AlertDescription>{form.error}</AlertDescription>
					</Alert>
				{/if}

				<div class="flex justify-between">
					<Button variant="ghost" href="/admin/staff?scope={data.scopeNodeId}">Cancel</Button>
					<Button type="submit">Create staff</Button>
				</div>
			</form>
		</Card.Content>
	</Card.Root>
</div>
